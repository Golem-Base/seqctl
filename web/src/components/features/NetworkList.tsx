import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useNetworkStore } from "@/stores/networkStore";
import { Activity, Network, RefreshCw, Server } from "lucide-react";
import { cn } from "@/lib/utils";

export function NetworkList() {
  const navigate = useNavigate();
  const {
    networks,
    loading,
    error,
    fetchNetworks,
    startAutoRefresh,
    stopAutoRefresh,
  } = useNetworkStore();

  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const [updatedNetworks, setUpdatedNetworks] = useState<Set<string>>(
    new Set(),
  );
  const previousNetworksRef = useRef<typeof networks>([]);

  useEffect(() => {
    // Initial fetch
    fetchNetworks();

    // Start auto-refresh
    startAutoRefresh();

    // Cleanup
    return () => {
      stopAutoRefresh();
    };
  }, []);

  // Update last updated time and detect changes
  useEffect(() => {
    if (networks.length > 0) {
      setLastUpdated(new Date());

      // Detect which networks have changed
      const changedNetworkIds = new Set<string>();

      networks.forEach((network) => {
        const previousNetwork = previousNetworksRef.current.find((n) =>
          n.id === network.id
        );

        if (
          !previousNetwork ||
          previousNetwork.healthy !== network.healthy ||
          previousNetwork.sequencers.length !== network.sequencers.length ||
          previousNetwork.sequencers.filter((s) => s.sequencer_active)
              .length !==
            network.sequencers.filter((s) => s.sequencer_active).length ||
          previousNetwork.sequencers.find((s) => s.conductor_leader)?.id !==
            network.sequencers.find((s) => s.conductor_leader)?.id
        ) {
          changedNetworkIds.add(network.id);
        }
      });

      // Update the set of changed networks
      if (
        changedNetworkIds.size > 0 && previousNetworksRef.current.length > 0
      ) {
        setUpdatedNetworks(changedNetworkIds);

        // Clear the animation after 2 seconds
        setTimeout(() => {
          setUpdatedNetworks(new Set());
        }, 2000);
      }

      // Store current networks for next comparison
      previousNetworksRef.current = networks;
    }
  }, [networks]);

  const handleRefresh = () => {
    fetchNetworks();
  };

  if (loading && networks.length === 0) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <RefreshCw className="h-8 w-8 animate-spin mx-auto mb-4 text-muted-foreground" />
          <p className="text-muted-foreground">Loading networks...</p>
        </div>
      </div>
    );
  }

  if (error && networks.length === 0) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <p className="text-destructive mb-4">{error}</p>
          <Button onClick={handleRefresh} variant="outline">
            <RefreshCw className="mr-2 h-4 w-4" />
            Retry
          </Button>
        </div>
      </div>
    );
  }

  if (networks.length === 0) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <Network className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
          <p className="text-muted-foreground mb-4">No networks found</p>
          <Button onClick={handleRefresh} variant="outline">
            <RefreshCw className="mr-2 h-4 w-4" />
            Refresh
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">Networks</h2>
        <div className="flex items-center gap-4">
          {lastUpdated && (
            <span className="text-sm text-muted-foreground">
              Last updated: {lastUpdated.toLocaleTimeString()}
            </span>
          )}
          <Button
            onClick={handleRefresh}
            variant="outline"
            size="sm"
            disabled={loading}
          >
            <RefreshCw
              className={`mr-2 h-4 w-4 ${loading ? "animate-spin" : ""}`}
            />
            Refresh
          </Button>
        </div>
      </div>

      {error && (
        <div className="bg-destructive/15 text-destructive px-4 py-2 rounded-md mb-4">
          {error}
        </div>
      )}

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {networks.map((network) => {
          const activeSequencers = network.sequencers.filter((s) =>
            s.sequencer_active
          ).length;
          const totalSequencers = network.sequencers.length;
          const leaderSequencer = network.sequencers.find((s) =>
            s.conductor_leader
          );

          return (
            <Card
              key={network.id}
              className={cn(
                "cursor-pointer transition-all hover:shadow-lg hover:scale-[1.02] hover:border-primary/50",
                updatedNetworks.has(network.id) && "animate-pulse-glow",
              )}
              onClick={() => navigate(`/networks/${network.id}`)}
            >
              <CardHeader>
                <div className="flex justify-between items-start">
                  <div>
                    <CardTitle className="flex items-center gap-2">
                      <Network className="h-5 w-5" />
                      {network.name}
                    </CardTitle>
                    <CardDescription className="mt-1">
                      {totalSequencers}{" "}
                      sequencer{totalSequencers !== 1 ? "s" : ""}
                    </CardDescription>
                  </div>
                  <Badge variant={network.healthy ? "success" : "destructive"}>
                    {network.healthy ? "Healthy" : "Unhealthy"}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-2 text-sm">
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">
                      Active Sequencers:
                    </span>
                    <span className="font-medium">
                      {activeSequencers} / {totalSequencers}
                    </span>
                  </div>

                  {leaderSequencer && (
                    <div className="flex items-center justify-between">
                      <span className="text-muted-foreground">Leader:</span>
                      <span className="font-medium flex items-center gap-1">
                        <Activity className="h-3 w-3" />
                        {leaderSequencer.id}
                      </span>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>
    </div>
  );
}
