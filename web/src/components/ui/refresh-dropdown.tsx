import { Clock, Pause, RefreshCw, Timer } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useNetworkStore } from "@/stores/networkStore";

interface RefreshDropdownProps {
  value: number;
  onChange: (value: number) => void;
}

const refreshOptions = [
  { value: 0, label: "Disabled", icon: Pause },
  { value: 5000, label: "5 seconds", icon: Timer },
  { value: 10000, label: "10 seconds", icon: Timer },
  { value: 30000, label: "30 seconds", icon: Timer },
  { value: 60000, label: "1 minute", icon: Clock },
];

export function RefreshDropdown({ value, onChange }: RefreshDropdownProps) {
  const isLoading = useNetworkStore((state) => state.loading);
  const isPaused = value === 0;

  const formatInterval = (milliseconds: number) => {
    if (milliseconds === 0) return "";
    const seconds = milliseconds / 1000;
    if (seconds < 60) {
      return `${seconds}s`;
    }
    return "1m";
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          size={isPaused ? "icon" : "sm"}
          className="relative h-9 px-2"
        >
          {isPaused ? <Pause className="h-4 w-4" /> : (
            <>
              <div className="flex items-center gap-1.5">
                <span className="text-xs font-medium text-muted-foreground">
                  {formatInterval(value)}
                </span>
                <RefreshCw
                  className="h-4 w-4"
                  style={{
                    animation: isLoading ? "spin 1s linear infinite" : "none",
                  }}
                />
              </div>
              <span className="absolute -bottom-1 -right-1 flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75">
                </span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500">
                </span>
              </span>
            </>
          )}
          <span className="sr-only">Auto-refresh settings</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-48">
        <DropdownMenuLabel>Auto-refresh interval</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuRadioGroup
          value={value.toString()}
          onValueChange={(v) => onChange(Number(v))}
        >
          {refreshOptions.map((option) => {
            const Icon = option.icon;
            return (
              <DropdownMenuRadioItem
                key={option.value}
                value={option.value.toString()}
                className="flex items-center gap-2"
              >
                <Icon className="h-3.5 w-3.5 text-muted-foreground" />
                <span>{option.label}</span>
              </DropdownMenuRadioItem>
            );
          })}
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
