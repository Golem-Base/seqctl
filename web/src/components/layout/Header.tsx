import { useNavigate, useParams } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { ThemeToggle } from "@/components/ui/theme-toggle";
import { RefreshDropdown } from "@/components/ui/refresh-dropdown";
import { useNetworkStore } from "@/stores/networkStore";
import { Activity, Settings } from "lucide-react";

export function Header() {
  const navigate = useNavigate();
  const { networkId } = useParams<{ networkId: string }>();
  const { setRefreshInterval, refreshInterval, networks } = useNetworkStore();

  const selectedNetwork = networks.find((n) => n.id === networkId);

  const handleHome = () => {
    navigate("/");
  };

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container mx-auto px-4">
        <div className="flex h-16 items-center justify-between">
          <div className="flex items-center gap-2">
            <button
              onClick={handleHome}
              className="text-xl font-bold flex items-center gap-2 hover:opacity-80 transition-opacity"
            >
              <Activity className="h-5 w-5" />
              Seqctl
            </button>

            {selectedNetwork && (
              <span className="text-muted-foreground">
                / {selectedNetwork.name}
              </span>
            )}
          </div>

          <div className="flex items-center gap-2">
            <RefreshDropdown
              value={refreshInterval}
              onChange={setRefreshInterval}
            />

            <ThemeToggle />

            <Button variant="ghost" size="icon">
              <Settings className="h-5 w-5" />
            </Button>
          </div>
        </div>
      </div>
    </header>
  );
}
