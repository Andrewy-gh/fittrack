import { useFieldContext } from '@/hooks/form';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

export default function NumberInput({ label }: { label: string }) {
  const field = useFieldContext<number>();

  return (
    <div className="space-y-1.5">
      <Label className="text-xs text-neutral-400">{label}</Label>
      <Input
        type="number"
        value={field.state.value || ''}
        onChange={(e) => field.handleChange(Number(e.target.value) || 0)}
        placeholder="0"
        className="bg-neutral-700 border-neutral-600 text-white text-center font-mono h-9"
      />
    </div>
  );
}
