import type { CurrentInternalUser, CurrentUser } from "@stackframe/react";
import { CustomUserButton } from "@/components/custom-user-button";
import { GuestUserButton } from "@/components/guest-user-button";

interface AccountSlotProps {
  user: CurrentUser | CurrentInternalUser | null;
}

export function AccountSlot({ user }: AccountSlotProps) {
  return user ? <CustomUserButton /> : <GuestUserButton />;
}
