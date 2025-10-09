export function MiniChart({
  data,
  activeIndex = -1,
}: {
  data: number[];
  activeIndex?: number;
}) {
  const maxValue = Math.max(...data);

  return (
    <div className="flex items-end gap-0.5 h-10">
      {data.map((value, index) => (
        <div
          key={index}
          className={`w-1 bg-muted rounded-sm transition-all duration-200 ${
            index === activeIndex ? 'bg-primary' : ''
          }`}
          style={{ height: `${(value / maxValue) * 100}%` }}
        />
      ))}
    </div>
  );
}
