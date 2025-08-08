import { useFieldContext } from '@/hooks/form';
import { cn } from '@/lib/utils';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

export default function SetTypeSelect({ className }: { className?: string }) {
  const field = useFieldContext<string>();
  return (
    <div className="space-y-4">
      <Label>Set Type</Label>
      <div>
        <Select value={field.state.value} onValueChange={field.handleChange}>
          <SelectTrigger
            className={cn(
              'text-sm h-9 w-full',
              'focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]',
              className
            )}
          >
            <SelectValue placeholder="Select type" />
          </SelectTrigger>
          <SelectContent position="popper" className="w-full">
            <SelectGroup>
              <SelectLabel>
                Set Type
              </SelectLabel>
              <SelectItem value="warmup" className="text-sm cursor-pointer">
                Warmup
              </SelectItem>
              <SelectItem value="working" className="text-sm cursor-pointer">
                Working
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>
    </div>
  );
}
