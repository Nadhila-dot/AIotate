import { useState, useRef, useCallback } from 'react';
import { 
  getTagsQuery, 
  getSubjectQuery, 
  getCourseQuery, 
  getDescriptionQuery,
  SheetResponse
} from "@/scripts/ai/getSheetinfo";

export interface FormData {
  subject: string;
  course: string;
  description: string;
  tags: string;
  visibility: string;
  curriculum: string;
  specialInstructions: string;
  mode: string;
  webSearchQuery?: string;
  webSearchEnabled?: boolean;
}

export interface LoadingState {
  tags: boolean;
  subject: boolean;
  course: boolean;
  description: boolean;
}

export interface CanRequestState {
  tags: boolean;
  subject: boolean;
  course: boolean;
  description: boolean;
}

export const useFormGeneration = (formData: FormData, setFormData: React.Dispatch<React.SetStateAction<FormData>>) => {
  const [autoTags, setAutoTags] = useState<string[]>([]);
  const [loadingState, setLoadingState] = useState<LoadingState>({
    tags: false,
    subject: false,
    course: false,
    description: false,
  });
  
  const [canRequestState, setCanRequestState] = useState<CanRequestState>({
    tags: true,
    subject: true,
    course: true,
    description: true,
  });
  
  const cooldownRefs = {
    tags: useRef<NodeJS.Timeout | null>(null),
    subject: useRef<NodeJS.Timeout | null>(null),
    course: useRef<NodeJS.Timeout | null>(null),
    description: useRef<NodeJS.Timeout | null>(null),
  };

  const processTags = useCallback((tags: string[] | undefined) => {
    if (tags && tags.length > 0) {
      setAutoTags(tags);
      setFormData(prev => ({
        ...prev,
        tags: tags.join(", ")
      }));
    }
  }, [setFormData]);

  const startCooldown = useCallback((field: keyof CanRequestState) => {
    if (cooldownRefs[field].current) clearTimeout(cooldownRefs[field].current);
    cooldownRefs[field].current = setTimeout(() => {
      setCanRequestState(prev => ({ ...prev, [field]: true }));
    }, 10000);
  }, []);

  const generateSubject = useCallback(() => {
    if (
      formData.subject.trim().length === 0 &&
      formData.course.trim().length >= 5 &&
      canRequestState.subject
    ) {
      setLoadingState(prev => ({ ...prev, subject: true }));
      setCanRequestState(prev => ({ ...prev, subject: false }));

      getSubjectQuery({
        course: formData.course,
        description: formData.description.trim() 
          ? formData.description 
          : "User didn't specify a description",
        generateTags: true
      })
        .then((response: SheetResponse) => {
          if (response.subject) {
            setFormData(prev => ({
              ...prev,
              subject: response.subject || ""
            }));
          }
          processTags(response.tags);
        })
        .finally(() => {
          setLoadingState(prev => ({ ...prev, subject: false }));
          startCooldown('subject');
        });
    }
  }, [formData, canRequestState.subject, processTags, setFormData, startCooldown]);

  const generateCourse = useCallback(() => {
    if (
      formData.course.trim().length === 0 &&
      formData.subject.trim().length >= 5 &&
      canRequestState.course
    ) {
      setLoadingState(prev => ({ ...prev, course: true }));
      setCanRequestState(prev => ({ ...prev, course: false }));

      getCourseQuery({
        subject: formData.subject,
        description: formData.description.trim() 
          ? formData.description 
          : "User didn't specify a description",
        generateTags: true
      })
        .then((response: SheetResponse) => {
          if (response.course) {
            setFormData(prev => ({
              ...prev,
              course: response.course || ""
            }));
          }
          processTags(response.tags);
        })
        .finally(() => {
          setLoadingState(prev => ({ ...prev, course: false }));
          startCooldown('course');
        });
    }
  }, [formData, canRequestState.course, processTags, setFormData, startCooldown]);

  const generateDescription = useCallback(() => {
    if (
      formData.description.trim().length === 0 &&
      formData.subject.trim().length >= 5 &&
      formData.course.trim().length >= 5 &&
      canRequestState.description
    ) {
      setLoadingState(prev => ({ ...prev, description: true }));
      setCanRequestState(prev => ({ ...prev, description: false }));
      
      getDescriptionQuery({
        subject: formData.subject,
        course: formData.course,
        generateTags: true
      })
        .then((response: SheetResponse) => {
          if (response.description) {
            setFormData(prev => ({
              ...prev,
              description: response.description || ""
            }));
          }
          processTags(response.tags);
        })
        .finally(() => {
          setLoadingState(prev => ({ ...prev, description: false }));
          startCooldown('description');
        });
    }
  }, [formData, canRequestState.description, processTags, setFormData, startCooldown]);

  const generateTags = useCallback(() => {
    if (
      formData.tags.trim().length === 0 &&
      formData.subject.trim().length >= 5 &&
      formData.course.trim().length >= 5 &&
      canRequestState.tags
    ) {
      setLoadingState(prev => ({ ...prev, tags: true }));
      setCanRequestState(prev => ({ ...prev, tags: false }));

      getTagsQuery({
        subject: formData.subject,
        course: formData.course,
        description: formData.description.trim() 
          ? formData.description 
          : "User didn't specify a description"
      })
        .then((tags: string[]) => {
          processTags(tags);
        })
        .finally(() => {
          setLoadingState(prev => ({ ...prev, tags: false }));
          startCooldown('tags');
        });
    }
  }, [formData, canRequestState.tags, processTags, startCooldown]);

  const generateMissingFields = useCallback(() => {
    if (formData.subject.trim().length === 0 && formData.course.trim().length >= 5) {
      generateSubject();
    }
    
    if (formData.course.trim().length === 0 && formData.subject.trim().length >= 5) {
      generateCourse();
    }
    
    if (formData.description.trim().length === 0 && 
        formData.subject.trim().length >= 5 && 
        formData.course.trim().length >= 5) {
      generateDescription();
    }
    
    if (formData.tags.trim().length === 0 &&
        formData.subject.trim().length >= 5 &&
        formData.course.trim().length >= 5) {
      generateTags();
    }
  }, [formData, generateSubject, generateCourse, generateDescription, generateTags]);

  return {
    autoTags,
    loadingState,
    canRequestState,
    generateSubject,
    generateCourse,
    generateDescription,
    generateTags,
    generateMissingFields,
  };
};