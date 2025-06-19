import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

export function SetTypeSelect({
  value,
  onChange,
}: {
  value: string;
  onChange: (value: string) => void;
}) {
  return (
    <Select value={value} onValueChange={onChange}>
      <SelectTrigger className="w-[180px]">
        <SelectValue placeholder="Select a set type" />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          <SelectLabel>Set Type</SelectLabel>
          <SelectItem value="warmup">Warmup</SelectItem>
          <SelectItem value="working">Working</SelectItem>
        </SelectGroup>
      </SelectContent>
    </Select>
  );
}
