import { Button } from '@/components/ui/button'
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { useFormContext } from 'react-hook-form'
import { X } from 'lucide-react'
import { z } from 'zod'
import React from 'react'

// Schema for the UpdateEtcdRaftOptionsPayload
export const updateEtcdRaftOptionsSchema = z.object({
	election_tick: z.number().int().positive('Election tick must be a positive integer'),
	heartbeat_tick: z.number().int().positive('Heartbeat tick must be a positive integer'),
	max_inflight_blocks: z.number().int().positive('Max inflight blocks must be a positive integer'),
	snapshot_interval_size: z.number().int().positive('Snapshot interval size must be a positive integer'),
	tick_interval: z.string().min(1, 'Tick interval is required'),
})

export type UpdateEtcdRaftOptionsFormValues = z.infer<typeof updateEtcdRaftOptionsSchema>

interface UpdateEtcdRaftOptionsOperationProps {
	index: number
	onRemove: () => void
	channelConfig?: {
		etcd_raft?: {
			election_tick?: number
			heartbeat_tick?: number
			max_inflight_blocks?: number
			snapshot_interval_size?: number
			tick_interval?: string
		}
	}
}

export function UpdateEtcdRaftOptionsOperation({ index, onRemove, channelConfig }: UpdateEtcdRaftOptionsOperationProps) {
	const form = useFormContext()

	// Set the default values from channel config if available
	React.useEffect(() => {
		if (channelConfig?.etcd_raft) {
			const { election_tick, heartbeat_tick, max_inflight_blocks, snapshot_interval_size, tick_interval } = channelConfig.etcd_raft
			if (election_tick) form.setValue(`operations.${index}.payload.election_tick`, election_tick)
			if (heartbeat_tick) form.setValue(`operations.${index}.payload.heartbeat_tick`, heartbeat_tick)
			if (max_inflight_blocks) form.setValue(`operations.${index}.payload.max_inflight_blocks`, max_inflight_blocks)
			if (snapshot_interval_size) form.setValue(`operations.${index}.payload.snapshot_interval_size`, snapshot_interval_size)
			if (tick_interval) form.setValue(`operations.${index}.payload.tick_interval`, tick_interval)
		}
	}, [channelConfig?.etcd_raft, form, index])

	return (
		<div className="rounded-lg border p-4 space-y-4">
			<div className="flex items-center justify-between">
				<h3 className="font-medium">Update Etcd Raft Options</h3>
				<Button type="button" variant="ghost" size="icon" onClick={onRemove}>
					<X className="h-4 w-4" />
				</Button>
			</div>
			<FormField
				control={form.control}
				name={`operations.${index}.payload.election_tick`}
				render={({ field }) => (
					<FormItem>
						<FormLabel>Election Tick</FormLabel>
						<FormControl>
							<Input type="number" placeholder="Enter election tick" {...field} onChange={(e) => field.onChange(parseInt(e.target.value) || 0)} />
						</FormControl>
						<FormMessage />
					</FormItem>
				)}
			/>
			<FormField
				control={form.control}
				name={`operations.${index}.payload.heartbeat_tick`}
				render={({ field }) => (
					<FormItem>
						<FormLabel>Heartbeat Tick</FormLabel>
						<FormControl>
							<Input type="number" placeholder="Enter heartbeat tick" {...field} onChange={(e) => field.onChange(parseInt(e.target.value) || 0)} />
						</FormControl>
						<FormMessage />
					</FormItem>
				)}
			/>
			<FormField
				control={form.control}
				name={`operations.${index}.payload.max_inflight_blocks`}
				render={({ field }) => (
					<FormItem>
						<FormLabel>Max Inflight Blocks</FormLabel>
						<FormControl>
							<Input type="number" placeholder="Enter max inflight blocks" {...field} onChange={(e) => field.onChange(parseInt(e.target.value) || 0)} />
						</FormControl>
						<FormMessage />
					</FormItem>
				)}
			/>
			<FormField
				control={form.control}
				name={`operations.${index}.payload.snapshot_interval_size`}
				render={({ field }) => (
					<FormItem>
						<FormLabel>Snapshot Interval Size</FormLabel>
						<FormControl>
							<Input type="number" placeholder="Enter snapshot interval size" {...field} onChange={(e) => field.onChange(parseInt(e.target.value) || 0)} />
						</FormControl>
						<FormMessage />
					</FormItem>
				)}
			/>
			<FormField
				control={form.control}
				name={`operations.${index}.payload.tick_interval`}
				render={({ field }) => (
					<FormItem>
						<FormLabel>Tick Interval</FormLabel>
						<FormControl>
							<Input placeholder="Enter tick interval (e.g. 500ms)" {...field} />
						</FormControl>
						<FormMessage />
					</FormItem>
				)}
			/>
		</div>
	)
} 