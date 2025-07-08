import { useFieldContext } from "@/hooks/form";
import { Label } from "@/components/ui/label";
import {Textarea} from "@/components/ui/textarea";

export default function NotesTextarea() {
    const field = useFieldContext<string>();
    return (
      <div className="space-y-3">
        <Label
          htmlFor={field.name}
          className="text-xs text-neutral-400 tracking-wider"
        >
          SESSION NOTES
        </Label>
        <Textarea
          id={field.name}
          name={field.name}
          value={field.state.value}
          onBlur={field.handleBlur}
          onChange={(e) => field.handleChange(e.target.value)}
          className="bg-neutral-800 border-neutral-600 text-white placeholder-neutral-500 min-h-[80px]"
          placeholder="Enter workout notes, focus areas, or observations..."
        />
      </div>
    );
}