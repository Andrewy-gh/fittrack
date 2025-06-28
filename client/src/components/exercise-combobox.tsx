import { useEffect, useState } from 'react';

import { Check, ChevronsUpDown, CirclePlus } from 'lucide-react';

import { Button } from '@/components/ui/button';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { cn } from '@/lib/utils';

import type { ExerciseOption } from '@/lib/types';

interface ComboboxProps {
  options: ExerciseOption[];
  selected: ExerciseOption['name'];
  className?: string;
  placeholder?: string;
  disabled?: boolean;
  onChange: (option: ExerciseOption) => void;
  onCreate?: (label: ExerciseOption['name']) => void;
}

/**
 * Convert katakana to hiragana(Only Japanese)
 *
 * カタカナをひらがなに変換する
 */
function toHiragana(value: string) {
  return value.replace(/[\u30a1-\u30f6]/g, function (match: string) {
    return String.fromCharCode(match.charCodeAt(0) - 0x60);
  });
}

/**
 * CommandItem to create a new query content
 *
 * クエリの内容を新規作成するCommandItem
 */
function CommandAddItem({
  query,
  onCreate,
}: {
  query: string;
  onCreate: () => void;
}) {
  return (
    <div
      tabIndex={0}
      onClick={onCreate}
      onKeyDown={(event: React.KeyboardEvent<HTMLDivElement>) => {
        if (event.key === 'Enter') {
          onCreate();
        }
      }}
      className={cn(
        'flex w-full text-blue-500 cursor-pointer text-sm px-2 py-1.5 rounded-sm items-center focus:outline-none',
        'hover:bg-blue-200 focus:!bg-blue-200'
      )}
    >
      <CirclePlus className="mr-2 h-4 w-4" />
      Create "{query}"
    </div>
  );
}

export function ExerciseCombobox({
  options,
  selected,
  className,
  placeholder,
  disabled,
  onChange,
  onCreate,
}: ComboboxProps) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [canCreate, setCanCreate] = useState(true);
  useEffect(() => {
    // Cannot create a new query if it is empty or has already been created
    // Unlike search, case sensitive here.

    // クエリが空の場合、またはすでに作成済みの場合は新規作成できない
    // 検索と違いここでは大文字小文字は区別する
    const isAlreadyCreated = !options.some((option) => option.name === query);
    setCanCreate(!!(query && isAlreadyCreated));
  }, [query, options]);

  function handleSelect(option: ExerciseOption) {
    if (onChange) {
      onChange(option);
      setOpen(false);
      setQuery('');
    }
  }

  function handleCreate() {
    if (onCreate && query) {
      onCreate(query);
      setOpen(false);
      setQuery('');
    }
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          type="button"
          variant="outline"
          role="combobox"
          disabled={disabled ?? false}
          aria-expanded={open}
          className={cn(
            'w-full justify-between bg-neutral-800 font-normal text-white border-neutral-700 hover:bg-neutral-700 hover:text-white',
            className
          )}
        >
          {selected && selected.length > 0 ? (
            <div className="truncate">
              {options.find((item) => item.name === selected)?.name}
            </div>
          ) : (
            <div className="text-neutral-400">{placeholder ?? 'Select exercise...'}</div>
          )}
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent
        className="w-full p-0 border-neutral-700 bg-neutral-800"
        align="start"
      >
        <Command  
          filter={(value, search) => {
            const v = toHiragana(value.toLocaleLowerCase());
            const s = toHiragana(search.toLocaleLowerCase());
            if (v.includes(s)) return 1;
            return 0;
          }}
          className="bg-neutral-800 text-white"
        >
          <CommandInput
            placeholder="Search exercises..."
            className="text-white placeholder-neutral-400"
            value={query}
            onValueChange={setQuery}
          />
          <CommandList className="border-t border-neutral-700">
            <CommandEmpty className="py-2 px-3 text-sm text-neutral-400">
              No exercises found.
            </CommandEmpty>
            <CommandGroup className="p-1">
              {options.map((option) => (
                <CommandItem
                  key={option.id}
                  value={option.name}
                  onSelect={() => handleSelect(option)}
                  className="aria-selected:bg-neutral-700 aria-selected:text-white text-neutral-200 hover:bg-neutral-700 hover:text-white"
                >
                  <Check
                    className={cn(
                      'mr-2 h-4 w-4',
                      selected === option.name ? 'opacity-100' : 'opacity-0'
                    )}
                  />
                  {option.name}
                </CommandItem>
              ))}
              {canCreate && query && (
                <div className="p-1">
                  <CommandAddItem
                    query={query}
                    onCreate={handleCreate}
                  />
                </div>
              )}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
