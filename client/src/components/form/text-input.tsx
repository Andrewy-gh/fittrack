import { useFieldContext } from "@/hooks/form";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export default function TextInput({label, placeholder}: {label: string, placeholder?: string}){
    const field = useFieldContext<string>();
    return (
      <div className="space-y-2">
        <Label className="text-xs text-neutral-400 tracking-wider">
          {label}
        </Label>
        <Input
          id={field.name}
          name={field.name}
          value={field.state.value}
          onBlur={field.handleBlur}
          onChange={(e) => field.handleChange(e.target.value)}
          className="bg-neutral-800 border-neutral-600 text-white placeholder-neutral-500"
          placeholder={placeholder}
        />
      </div>
    );
}