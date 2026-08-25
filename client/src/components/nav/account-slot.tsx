import type { ApplicationUser } from "@/lib/application-user";
import { CustomUserButton } from "@/components/custom-user-button";
import { GuestUserButton } from "@/components/guest-user-button";

interface AccountSlotProps {
  user: ApplicationUser | null;
}

export function AccountSlot({ user }: AccountSlotProps) {
  return user ? <CustomUserButton /> : <GuestUserButton />;
}
