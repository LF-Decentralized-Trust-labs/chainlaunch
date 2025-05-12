import { useParams } from "react-router-dom";
import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { MetricsCard, MetricsDataPoint } from "@/components/metrics/MetricsCard";
import { MetricsGrid } from "@/components/metrics/MetricsGrid";
import { useMetrics } from "@/hooks/useMetrics";
import { getApiV1MetricsNodeById } from "@/api/client";
import { useQuery } from "@tanstack/react-query";
import { HttpNodeResponse } from "@/api/client/types.gen";
interface BesuMetricsPageProps {
	nodeId: number;
}
export default function BesuMetricsPage({ nodeId }: BesuMetricsPageProps) {
  const [timeRange, setTimeRange] = useState({ start: Date.now() - 3600000, end: Date.now() });

  const { data: nodeData } = useQuery({
    queryKey: ['node', nodeId],
    queryFn: async () => {
      if (!nodeId) return null;
      const response = await getApiV1MetricsNodeById({ path: { id: nodeId.toString() } });
      return response.data as HttpNodeResponse;
    },
    enabled: !!nodeId
  });

  // System metrics
  const { data: cpuData } = useMetrics({
    nodeId: nodeId || "",
    query: 'rate(process_cpu_seconds_total[1m])',
    start: timeRange.start,
    end: timeRange.end,
    step: 60
  });

  const { data: memoryData } = useMetrics({
    nodeId: nodeId || "",
    query: 'process_resident_memory_bytes',
    start: timeRange.start,
    end: timeRange.end,
    step: 60
  });

  const { data: diskData } = useMetrics({
    nodeId: nodeId || "",
    query: 'node_filesystem_size_bytes{mountpoint="/"} - node_filesystem_free_bytes{mountpoint="/"}',
    start: timeRange.start,
    end: timeRange.end,
    step: 60
  });

  // Besu specific metrics
  const { data: blockHeightData } = useMetrics({
    nodeId: nodeId || "",
    query: 'besu_blockchain_height',
    start: timeRange.start,
    end: timeRange.end,
    step: 60
  });

  const { data: transactionCountData } = useMetrics({
    nodeId: nodeId || "",
    query: 'rate(besu_transactions_total[1m])',
    start: timeRange.start,
    end: timeRange.end,
    step: 60
  });

  const { data: gasUsedData } = useMetrics({
    nodeId: nodeId || "",
    query: 'rate(besu_blockchain_gas_used[1m])',
    start: timeRange.start,
    end: timeRange.end,
    step: 60
  });

  const { data: peerCountData } = useMetrics({
    nodeId: nodeId || "",
    query: 'besu_network_peers',
    start: timeRange.start,
    end: timeRange.end,
    step: 60
  });

  const { data: pendingTransactionsData } = useMetrics({
    nodeId: nodeId || "",
    query: 'besu_pending_transactions',
    start: timeRange.start,
    end: timeRange.end,
    step: 60
  });

  const { data: syncStatusData } = useMetrics({
    nodeId: nodeId || "",
    query: 'besu_sync_status',
    start: timeRange.start,
    end: timeRange.end,
    step: 60
  });

  const { data: blockPropagationTimeData } = useMetrics({
    nodeId: nodeId || "",
    query: 'besu_block_propagation_time_seconds',
    start: timeRange.start,
    end: timeRange.end,
    step: 60
  });

  if (!nodeId || !nodeData) {
    return null;
  }

  return (
    <div className="container mx-auto py-6">
      <h1 className="text-2xl font-bold mb-6">Besu Metrics for {nodeData.name}</h1>
      
      <MetricsGrid>
        <MetricsCard
          title="CPU Usage"
          data={cpuData || []}
          color="#2563eb"
          unit="%"
          valueFormatter={(value: number) => `${(value * 100).toFixed(2)}%`}
        />
        
        <MetricsCard
          title="Memory Usage"
          data={memoryData || []}
          color="#16a34a"
          unit="bytes"
          valueFormatter={(value: number) => `${(value / 1024 / 1024).toFixed(2)} MB`}
        />
        
        <MetricsCard
          title="Disk Usage"
          data={diskData || []}
          color="#dc2626"
          unit="bytes"
          valueFormatter={(value: number) => `${(value / 1024 / 1024 / 1024).toFixed(2)} GB`}
        />

        <MetricsCard
          title="Block Height"
          data={blockHeightData || []}
          color="#9333ea"
          unit="blocks"
          valueFormatter={(value: number) => value.toFixed(0)}
        />

        <MetricsCard
          title="Transaction Rate"
          data={transactionCountData || []}
          color="#eab308"
          unit="tx/s"
          valueFormatter={(value: number) => value.toFixed(2)}
        />

        <MetricsCard
          title="Gas Used Rate"
          data={gasUsedData || []}
          color="#0891b2"
          unit="gas/s"
          valueFormatter={(value: number) => value.toFixed(0)}
        />

        <MetricsCard
          title="Peer Count"
          data={peerCountData || []}
          color="#be185d"
          unit="peers"
          valueFormatter={(value: number) => value.toFixed(0)}
        />

        <MetricsCard
          title="Pending Transactions"
          data={pendingTransactionsData || []}
          color="#dc2626"
          unit="tx"
          valueFormatter={(value: number) => value.toFixed(0)}
        />

        <MetricsCard
          title="Sync Status"
          data={syncStatusData || []}
          color="#16a34a"
          unit="%"
          valueFormatter={(value: number) => `${(value * 100).toFixed(2)}%`}
        />

        <MetricsCard
          title="Block Propagation Time"
          data={blockPropagationTimeData || []}
          color="#9333ea"
          unit="seconds"
          valueFormatter={(value: number) => value.toFixed(3)}
        />
      </MetricsGrid>
    </div>
  );
} 