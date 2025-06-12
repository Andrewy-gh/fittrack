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
          className={cn('w-80 font-normal', className)}
        >
          {selected && selected.length > 0 ? (
            <div className="truncate mr-auto">
              {options.find((item) => item.name === selected)?.name}
            </div>
          ) : (
            <div className="text-slate-600 mr-auto">
              {placeholder ?? 'Select'}
            </div>
          )}
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80 p-0" align="start">
        <Command
          filter={(value, search) => {
            const v = toHiragana(value.toLocaleLowerCase());
            const s = toHiragana(search.toLocaleLowerCase());
            if (v.includes(s)) return 1;
            return 0;
          }}
        >
          <CommandInput
            placeholder="Search or create new"
            value={query}
            onValueChange={(value: string) => setQuery(value)}
            onKeyDown={(event: React.KeyboardEvent<HTMLInputElement>) => {
              if (event.key === 'Enter') {
                // Avoid selecting what is displayed as a choice even if you press Enter for the conversion
                // Note that if you do this, you can select a choice with the arrow keys and press Enter, but it will not be selected

                // 変換のためにEnterを押した場合でも、選択肢として表示されているものが選択されてしまうのを回避する
                // この処理をする場合、選択肢を矢印キーで選択してEnterを押しても選択されなくなるので注意
                event.preventDefault();
              }
            }}
          />
          <CommandEmpty className="flex pl-1 py-1 w-full">
            {query && (
              <CommandAddItem query={query} onCreate={() => handleCreate()} />
            )}
          </CommandEmpty>

          <CommandList>
            <CommandGroup className="overflow-y-auto">
              {/* No options and no query */}
              {/* Even if written as a Child of CommandEmpty, it may not be displayed only the first time, so write it in CommandGroup. */}

              {/* 選択肢が1つも無い、かつクエリも入力されていない場合 */}
              {/* CommandEmptyのChildとして書いても、初回だけ表示されない場合があるのでCommandGroup内に書く */}
              {options.length === 0 && !query && (
                <div className="py-1.5 pl-8 space-y-1 text-sm">
                  <p>No items</p>
                  <p>Enter a value to create a new one</p>
                </div>
              )}

              {/* Create */}
              {canCreate && (
                <CommandAddItem query={query} onCreate={() => handleCreate()} />
              )}

              {/* Select */}
              {options.map((option, index) => (
                <CommandItem
                  key={option.name}
                  tabIndex={0}
                  value={option.name}
                  onSelect={() => {
                    console.log('onSelect');
                    handleSelect(option);
                  }}
                  onKeyDown={(event: React.KeyboardEvent<HTMLDivElement>) => {
                    if (event.key === 'Enter') {
                      // Process to prevent onSelect from firing, but it does not work well with StackBlitz.
                      // onSelectの発火を防ぐ処理だが、StackBlitzだとうまく動作しない
                      event.stopPropagation();

                      handleSelect(option);
                    }
                  }}
                  className={cn(
                    'cursor-pointer',
                    // Override CommandItem class name
                    // CommandItemのクラス名を上書き
                    'focus:!bg-blue-200 hover:!bg-blue-200 aria-selected:bg-transparent'
                  )}
                >
                  {/* min to avoid the check icon being too small when the option.label is long. */}
                  {/* minを設定していない場合、option.labelが長いとチェックアイコンが小さくなってしまう */}
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
      </PopoverContent>
    </Popover>
  );
}
