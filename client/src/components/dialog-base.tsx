import { type ReactNode } from 'react';
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';

export function DialogBase({
  isOpen,
  description,
  children,
}: {
  isOpen: boolean;
  description: ReactNode;
  children: ReactNode;
}) {
  return (
    <AlertDialog open={isOpen}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
          <AlertDialogDescription>{description}</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>{children}</AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
