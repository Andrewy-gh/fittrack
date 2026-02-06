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

interface ComboboxProps<T extends { name: string }> {
  /** Array of options to display in the combobox */
  options: T[];
  /** The currently selected option's name */
  selected: string;
  /** Additional CSS classes to apply to the trigger button */
  className?: string;
  /** Accessible label for the combobox trigger */
  ariaLabel?: string;
  /** Accessible label for the search input */
  inputAriaLabel?: string;
  /** Placeholder text to display when no option is selected */
  placeholder?: string;
  /** Whether the combobox is disabled */
  disabled?: boolean;
  /** Callback function called when an option is selected */
  onChange: (option: T) => void;
  /** Callback function called when a new option is created */
  onCreate?: (label: string) => void;
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
      role="option"
      aria-label={`Create "${query}"`}
      tabIndex={0}
      onClick={onCreate}
      onKeyDown={(event: KeyboardEvent<HTMLDivElement>) => {
        if (event.key === 'Enter') {
          onCreate();
        }
      }}
      className="flex w-full cursor-pointer items-center rounded-sm px-2 py-1.5 focus:outline-none"
    >
      <CirclePlus className="mr-2 h-4 w-4" />
      Create "{query}"
    </div>
  );
}

function GenericList<T extends { name: string }>({
  options,
  selected,
  query,
  setQuery,
  canCreate,
  setOpen,
  onChange,
  onCreate,
  inputAriaLabel,
}: {
  options: T[];
  selected: string;
  query: string;
  setQuery: (query: string) => void;
  canCreate: boolean;
  setOpen: (open: boolean) => void;
  onChange: (option: T) => void;
  onCreate?: (label: string) => void;
  inputAriaLabel?: string;
}) {
  function handleSelect(option: T) {
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
      className="pb-4"
    >
      <CommandInput
        placeholder="Search options..."
        aria-label={inputAriaLabel ?? 'Search options'}
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
      <CommandList>
        <CommandGroup className="p-1">
          {/* No options and no query */}
          {options.length === 0 && !query && (
            <div className="py-1.5 pl-8 space-y-1">
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
              key={option.name} // Use name as key
              value={option.name}
              tabIndex={0}
              onSelect={() => handleSelect(option)}
              onKeyDown={(event: KeyboardEvent<HTMLDivElement>) => {
                if (event.key === 'Enter') {
                  event.stopPropagation();
                  handleSelect(option);
                }
              }}
              className={cn('cursor-pointer')}
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

export function GenericCombobox<T extends { name: string }>({
  options,
  selected,
  className,
  ariaLabel,
  inputAriaLabel,
  placeholder,
  disabled,
  onChange,
  onCreate,
}: ComboboxProps<T>) {
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
      aria-label={ariaLabel}
      disabled={disabled ?? false}
      aria-expanded={open}
      className={cn('w-full justify-between font-normal', className)}
    >
      {selected && selected.length > 0 ? (
        <div className="truncate">
          {options.find((item) => item.name === selected)?.name}
        </div>
      ) : (
        <div>{placeholder ?? 'Select an option...'}</div>
      )}
      <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
    </Button>
  );

  if (isDesktop) {
    return (
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>{triggerButton}</PopoverTrigger>
        <PopoverContent className="w-full p-0" align="start">
          <GenericList
            options={options}
            selected={selected}
            query={query}
            setQuery={setQuery}
            canCreate={canCreate}
            setOpen={setOpen}
            onChange={onChange}
            onCreate={onCreate}
            inputAriaLabel={inputAriaLabel}
          />
        </PopoverContent>
      </Popover>
    );
  }

  return (
    <Drawer open={open} onOpenChange={setOpen}>
      <DrawerTrigger asChild>{triggerButton}</DrawerTrigger>
      <DrawerContent>
        <div className="mt-4">
          <GenericList
            options={options}
            selected={selected}
            query={query}
            setQuery={setQuery}
            canCreate={canCreate}
            setOpen={setOpen}
            onChange={onChange}
            onCreate={onCreate}
            inputAriaLabel={inputAriaLabel}
          />
        </div>
      </DrawerContent>
    </Drawer>
  );
}
