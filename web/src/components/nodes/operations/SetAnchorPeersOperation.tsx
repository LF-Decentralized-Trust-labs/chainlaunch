import { Button } from '@/components/ui/button'
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { useFormContext } from 'react-hook-form'
import { X } from 'lucide-react'
import { z } from 'zod'
import React from 'react'

export const setAnchorPeersSchema = z.object({
	msp_id: z.string().min(1, 'MSP ID is required'),
	anchor_peers: z.array(
		z.object({
			host: z.string().min(1, 'Host is required'),
			port: z.number().int().positive('Port must be a positive integer'),
		})
	).min(1, 'At least one anchor peer is required'),
})

interface AnchorPeer {
	host: string
	port: number
}

interface SetAnchorPeersOperationProps {
	index: number
	onRemove: () => void
	channelConfig?: {
		anchor_peers?: AnchorPeer[]
	}
}

export function SetAnchorPeersOperation({ index, onRemove, channelConfig }: SetAnchorPeersOperationProps) {
	const form = useFormContext()

	// Set the default values from channel config if available
	React.useEffect(() => {
		if (channelConfig?.anchor_peers) {
			form.setValue(`operations.${index}.payload.anchor_peers`, channelConfig.anchor_peers)
		}
	}, [channelConfig?.anchor_peers, form, index])

	return (
		<div className="rounded-lg border p-4 space-y-4">
			<div className="flex items-center justify-between">
				<h3 className="font-medium">Set Anchor Peers</h3>
				<Button type="button" variant="ghost" size="icon" onClick={onRemove}>
					<X className="h-4 w-4" />
				</Button>
			</div>
			<FormField
				control={form.control}
				name={`operations.${index}.payload.msp_id`}
				render={({ field }) => (
					<FormItem>
						<FormLabel>MSP ID</FormLabel>
						<FormControl>
							<Input placeholder="Enter MSP ID" {...field} />
						</FormControl>
						<FormMessage />
					</FormItem>
				)}
			/>
			<FormField
				control={form.control}
				name={`operations.${index}.payload.anchor_peers`}
				render={({ field }) => (
					<FormItem>
						<FormLabel>Anchor Peers</FormLabel>
						<FormControl>
							<div className="space-y-4">
								{(field.value as AnchorPeer[])?.map((peer: AnchorPeer, peerIndex: number) => (
									<div key={peerIndex} className="flex gap-4">
										<Input
											placeholder="Enter host"
											value={peer.host}
											onChange={(e) => {
												const newPeers = [...field.value as AnchorPeer[]]
												newPeers[peerIndex] = { ...newPeers[peerIndex], host: e.target.value }
												field.onChange(newPeers)
											}}
										/>
										<Input
											type="number"
											placeholder="Enter port"
											value={peer.port}
											onChange={(e) => {
												const newPeers = [...field.value as AnchorPeer[]]
												newPeers[peerIndex] = { ...newPeers[peerIndex], port: parseInt(e.target.value) || 0 }
												field.onChange(newPeers)
											}}
										/>
										<Button
											type="button"
											variant="ghost"
											size="icon"
											onClick={() => {
												const newPeers = (field.value as AnchorPeer[]).filter((_: AnchorPeer, i: number) => i !== peerIndex)
												field.onChange(newPeers)
											}}
										>
											<X className="h-4 w-4" />
										</Button>
									</div>
								))}
								<Button
									type="button"
									variant="outline"
									onClick={() => {
										field.onChange([...(field.value as AnchorPeer[]), { host: '', port: 7051 }])
									}}
								>
									Add Anchor Peer
								</Button>
							</div>
						</FormControl>
						<FormMessage />
					</FormItem>
				)}
			/>
		</div>
	)
} 