import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { cn } from '@/lib/utils';

export function SetTypeSelect({
  value,
  onChange,
  className,
}: {
  value: string;
  onChange: (value: string) => void;
  className?: string;
}) {
  return (
    <Select value={value} onValueChange={onChange}>
      <SelectTrigger 
        className={cn(
          'bg-neutral-700 border-neutral-600 text-white font-mono text-sm h-9',
          'hover:bg-neutral-600 focus:ring-1 focus:ring-orange-500 focus:ring-offset-1 focus:ring-offset-neutral-800',
          'transition-colors duration-200',
          className
        )}
      >
        <SelectValue placeholder="Select type" className="placeholder:text-neutral-400" />
      </SelectTrigger>
      <SelectContent 
        className="bg-neutral-800 border-neutral-700 text-white" 
        position="popper"
      >
        <SelectGroup>
          <SelectLabel className="text-xs text-neutral-400 px-2 py-1.5">SET TYPE</SelectLabel>
          <SelectItem 
            value="warmup" 
            className="text-sm focus:bg-neutral-700 focus:text-white cursor-pointer"
          >
            Warmup
          </SelectItem>
          <SelectItem 
            value="working" 
            className="text-sm focus:bg-neutral-700 focus:text-white cursor-pointer"
          >
            Working
          </SelectItem>
        </SelectGroup>
      </SelectContent>
    </Select>
  );
}
