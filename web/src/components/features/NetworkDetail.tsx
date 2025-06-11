import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Network, Sequencer } from "@/api/types";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { useNetworkStore } from "@/stores/networkStore";
import { useActionStore } from "@/stores/actionStore";
import {
  Activity,
  ArrowRightLeft,
  CheckCircle,
  Crown,
  MoreVertical,
  Pause,
  Play,
  Server,
  Shield,
  StopCircle,
  UserMinus,
  Users,
  Zap,
} from "lucide-react";

export function NetworkDetail() {
  const { networkId } = useParams<{ networkId: string }>();
  const navigate = useNavigate();
  const { networks, fetchNetwork, refreshInterval } = useNetworkStore();

  const network = networks.find((n) => n.id === networkId);

  useEffect(() => {
    if (!networkId) return;

    // Initial fetch of just this network
    fetchNetwork(networkId);

    // Set up refresh interval for just this network
    const timer = refreshInterval > 0
      ? setInterval(() => {
        fetchNetwork(networkId);
      }, refreshInterval)
      : null;

    // Cleanup
    return () => {
      if (timer) clearInterval(timer);
    };
  }, [networkId, fetchNetwork, refreshInterval]);

  useEffect(() => {
    if (networks.length > 0 && !network) {
      navigate("/");
    }
  }, [networks.length, network, navigate]);
  const { pauseSequencer, resumeSequencer, haltSequencer } = useNetworkStore();
  const { openModal } = useActionStore();
  const [loadingAction, setLoadingAction] = useState<string | null>(null);

  const handlePause = async (sequencerId: string) => {
    setLoadingAction(`pause-${sequencerId}`);
    try {
      await pauseSequencer(sequencerId);
    } finally {
      setLoadingAction(null);
    }
  };

  const handleResume = async (sequencerId: string) => {
    setLoadingAction(`resume-${sequencerId}`);
    try {
      await resumeSequencer(sequencerId);
    } finally {
      setLoadingAction(null);
    }
  };

  const handleHalt = async (sequencerId: string) => {
    setLoadingAction(`halt-${sequencerId}`);
    try {
      await haltSequencer(sequencerId);
    } finally {
      setLoadingAction(null);
    }
  };

  const getSequencerStatusIcon = (sequencer: Sequencer) => {
    if (!sequencer.conductor_active) {
      return <Pause className="h-4 w-4 text-muted-foreground" />;
    }
    if (sequencer.conductor_leader) {
      return <Crown className="h-4 w-4 text-yellow-500" />;
    }
    if (sequencer.sequencer_active) {
      return <Activity className="h-4 w-4 text-green-500" />;
    }
    return <CheckCircle className="h-4 w-4 text-blue-500" />;
  };

  const getHealthBadge = (sequencer: Sequencer) => {
    if (!sequencer.sequencer_healthy) {
      return <Badge variant="destructive">Unhealthy</Badge>;
    }
    if (sequencer.conductor_paused) {
      return <Badge variant="secondary">Paused</Badge>;
    }
    if (sequencer.conductor_stopped) {
      return <Badge variant="secondary">Stopped</Badge>;
    }
    return <Badge variant="success">Healthy</Badge>;
  };

  if (!network) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <p className="text-muted-foreground">Loading network details...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Network Header */}
      <div className="flex justify-between items-start">
        <div>
          <h1 className="text-3xl font-bold">{network.name}</h1>
          <p className="text-muted-foreground mt-1">
            Network overview and sequencer management
          </p>
        </div>
        <Badge
          variant={network.healthy ? "success" : "destructive"}
          className="text-sm"
        >
          {network.healthy ? "Healthy" : "Unhealthy"}
        </Badge>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Total Sequencers
            </CardTitle>
            <Server className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {network.sequencers.length}
            </div>
            <p className="text-xs text-muted-foreground">
              Sequencers in network
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Active Sequencers
            </CardTitle>
            <Activity className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {network.sequencers.filter((s) => s.sequencer_active).length}
            </div>
            <p className="text-xs text-muted-foreground">Currently active</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Voting Members
            </CardTitle>
            <Shield className="h-4 w-4 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {network.sequencers.filter((s) => s.voting).length}
            </div>
            <p className="text-xs text-muted-foreground">
              Can participate in consensus
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Sequencers Table */}
      <Card>
        <CardHeader>
          <CardTitle>Sequencers</CardTitle>
          <CardDescription>
            Manage and monitor sequencers in this network
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Status</TableHead>
                <TableHead>ID</TableHead>
                <TableHead>Unsafe L2 Block</TableHead>
                <TableHead>Health</TableHead>
                <TableHead>Role</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {network.sequencers.map((sequencer) => (
                <TableRow key={sequencer.id}>
                  <TableCell>{getSequencerStatusIcon(sequencer)}</TableCell>
                  <TableCell className="font-medium">
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger className="cursor-help">
                          {sequencer.id}
                        </TooltipTrigger>
                        <TooltipContent>
                          <p className="font-mono text-xs">
                            Raft: {sequencer.raft_addr}
                          </p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </TableCell>
                  <TableCell className="font-mono text-sm">
                    {sequencer.unsafe_l2.toLocaleString()}
                  </TableCell>
                  <TableCell>{getHealthBadge(sequencer)}</TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      {sequencer.conductor_leader && (
                        <Badge variant="outline" className="gap-1">
                          <Crown className="h-3 w-3" />
                          Leader
                        </Badge>
                      )}
                      {sequencer.voting && (
                        <Badge variant="outline">Voting</Badge>
                      )}
                    </div>
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex items-center gap-2 justify-end">
                      {/* Quick Actions */}
                      {sequencer.conductor_active
                        ? (
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => handlePause(sequencer.id)}
                            disabled={loadingAction === `pause-${sequencer.id}`}
                          >
                            <Pause className="h-4 w-4" />
                          </Button>
                        )
                        : (
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => handleResume(sequencer.id)}
                            disabled={loadingAction ===
                              `resume-${sequencer.id}`}
                          >
                            <Play className="h-4 w-4" />
                          </Button>
                        )}

                      {sequencer.sequencer_active && (
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => handleHalt(sequencer.id)}
                          disabled={loadingAction === `halt-${sequencer.id}`}
                        >
                          <StopCircle className="h-4 w-4" />
                        </Button>
                      )}

                      {/* More Actions Dropdown */}
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button size="sm" variant="ghost">
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="w-48">
                          {!sequencer.conductor_leader && (
                            <DropdownMenuItem
                              onClick={() =>
                                openModal("transfer-leader", sequencer.id)}
                            >
                              <ArrowRightLeft className="mr-2 h-4 w-4" />
                              Transfer Leader
                            </DropdownMenuItem>
                          )}

                          {sequencer.conductor_leader && (
                            <DropdownMenuItem
                              onClick={() =>
                                openModal("resign-leader", sequencer.id)}
                            >
                              <Crown className="mr-2 h-4 w-4" />
                              Resign Leader
                            </DropdownMenuItem>
                          )}

                          <DropdownMenuItem
                            onClick={() =>
                              openModal("override-leader", sequencer.id)}
                          >
                            <Shield className="mr-2 h-4 w-4" />
                            Override Leader
                          </DropdownMenuItem>

                          {!sequencer.sequencer_active && (
                            <DropdownMenuItem
                              onClick={() =>
                                openModal("force-active", sequencer.id)}
                            >
                              <Zap className="mr-2 h-4 w-4" />
                              Force Active
                            </DropdownMenuItem>
                          )}

                          {sequencer.conductor_leader && (
                            <>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                onClick={() =>
                                  openModal("update-member", sequencer.id)}
                              >
                                <Users className="mr-2 h-4 w-4" />
                                Update Membership
                              </DropdownMenuItem>

                              <DropdownMenuItem
                                onClick={() =>
                                  openModal("remove-member", sequencer.id)}
                                className="text-destructive focus:text-destructive"
                              >
                                <UserMinus className="mr-2 h-4 w-4" />
                                Remove Member
                              </DropdownMenuItem>
                            </>
                          )}
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
