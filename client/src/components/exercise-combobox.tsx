import { type KeyboardEvent, useEffect, useState } from 'react';
import { Check, ChevronsUpDown, CirclePlus } from 'lucide-react';
import { useMediaQuery } from '@/hooks/use-media-query';
import { Button } from '@/components/ui/button';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command';
import { Drawer, DrawerContent, DrawerTrigger } from '@/components/ui/drawer';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { cn } from '@/lib/utils';
import type { ExerciseOption } from '@/lib/api/exercises';

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
 * CommandItem to create a new query content
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
      onKeyDown={(event: KeyboardEvent<HTMLDivElement>) => {
        if (event.key === 'Enter') {
          onCreate();
        }
      }}
      className={cn(
        'flex w-full text-neutral-200 cursor-pointer text-sm px-2 py-1.5 rounded-sm items-center focus:outline-none',
        'hover:bg-neutral-700 focus:!bg-neutral-700'
      )}
    >
      <CirclePlus className="mr-2 h-4 w-4" />
      Create "{query}"
    </div>
  );
}

function ExerciseList({
  options,
  selected,
  query,
  setQuery,
  canCreate,
  setOpen,
  onChange,
  onCreate,
}: {
  options: ExerciseOption[];
  selected: ExerciseOption['name'];
  query: string;
  setQuery: (query: string) => void;
  canCreate: boolean;
  setOpen: (open: boolean) => void;
  onChange: (option: ExerciseOption) => void;
  onCreate?: (label: ExerciseOption['name']) => void;
}) {
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
    <Command
      filter={(value, search) => {
        const v = value.toLowerCase();
        const s = search.toLowerCase();
        if (v.includes(s)) return 1;
        return 0;
      }}
      className="bg-neutral-800 text-white"
    >
      <CommandInput
        placeholder="Search exercises..."
        className="text-white placeholder-neutral-400"
        value={query}
        onValueChange={(value: string) => setQuery(value)}
        onKeyDown={(event: KeyboardEvent<HTMLInputElement>) => {
          if (event.key === 'Enter') {
            // Avoid selecting what is displayed as a choice even if you press Enter for the conversion
            event.preventDefault();
          }
        }}
      />
      <CommandEmpty className="flex pl-1 py-1 w-full">
        {query && canCreate && (
          <CommandAddItem query={query} onCreate={handleCreate} />
        )}
      </CommandEmpty>
      <CommandList className="border-t border-neutral-700">
        <CommandGroup className="p-1">
          {/* No options and no query */}
          {options.length === 0 && !query && (
            <div className="py-1.5 pl-8 space-y-1 text-sm text-neutral-400">
              <p>No items</p>
              <p>Enter a value to create a new one</p>
            </div>
          )}
          {/* Create option - shown when there are existing options but query doesn't match */}
          {options.length > 0 && canCreate && (
            <CommandAddItem query={query} onCreate={handleCreate} />
          )}
          {/* Select options */}
          {options.map((option) => (
            <CommandItem
              key={option.id || option.name}
              value={option.name}
              tabIndex={0}
              onSelect={() => handleSelect(option)}
              onKeyDown={(event: KeyboardEvent<HTMLDivElement>) => {
                if (event.key === 'Enter') {
                  event.stopPropagation();
                  handleSelect(option);
                }
              }}
              className={cn(
                'cursor-pointer',
                'aria-selected:bg-neutral-700 aria-selected:text-white text-neutral-200 hover:bg-neutral-700 hover:text-white focus:!bg-neutral-700'
              )}
            >
              <Check
                className={cn(
                  'mr-2 h-4 w-4 min-h-4 min-w-4',
                  selected === option.name ? 'opacity-100' : 'opacity-0'
                )}
              />
              {option.name}
            </CommandItem>
          ))}
        </CommandGroup>
      </CommandList>
    </Command>
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
  const [canCreate, setCanCreate] = useState(false);
  const isDesktop = useMediaQuery('(min-width: 768px)');

  useEffect(() => {
    // Cannot create a new query if it is empty or has already been created
    const queryExists = options.some((option) => option.name === query);
    setCanCreate(!!(query && !queryExists));
  }, [query, options]);

  const triggerButton = (
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
        <div className="text-neutral-400">
          {placeholder ?? 'Select exercise...'}
        </div>
      )}
      <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
    </Button>
  );

  if (isDesktop) {
    return (
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>{triggerButton}</PopoverTrigger>
        <PopoverContent
          className="w-full p-0 border-neutral-700 bg-neutral-800"
          align="start"
        >
          <ExerciseList
            options={options}
            selected={selected}
            query={query}
            setQuery={setQuery}
            canCreate={canCreate}
            setOpen={setOpen}
            onChange={onChange}
            onCreate={onCreate}
          />
        </PopoverContent>
      </Popover>
    );
  }

  return (
    <Drawer open={open} onOpenChange={setOpen}>
      <DrawerTrigger asChild>{triggerButton}</DrawerTrigger>
      <DrawerContent className="bg-neutral-900 border-neutral-700">
        <div className="mt-4 border-t border-neutral-700">
          <ExerciseList
            options={options}
            selected={selected}
            query={query}
            setQuery={setQuery}
            canCreate={canCreate}
            setOpen={setOpen}
            onChange={onChange}
            onCreate={onCreate}
          />
        </div>
      </DrawerContent>
    </Drawer>
  );
}
