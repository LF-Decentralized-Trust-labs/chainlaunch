import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { zodResolver } from '@hookform/resolvers/zod'
import { Plus } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'
import { postNetworksFabricByIdNodesMutation } from '@/api/client/@tanstack/react-query.gen'
import { useMutation } from '@tanstack/react-query'

const formSchema = z.object({
	nodeId: z.string({
		required_error: 'Please select a node',
	}),
})

type FormValues = z.infer<typeof formSchema>

interface AddNodeDialogProps {
	networkId: number
	availableNodes: Array<{
		id: number
		name: string
		nodeType: string
	}>
	onNodeAdded: () => void
}

export function AddNodeDialog({ networkId, availableNodes, onNodeAdded }: AddNodeDialogProps) {
	const [open, setOpen] = useState(false)
	const [isLoading, setIsLoading] = useState(false)

	const form = useForm<FormValues>({
		resolver: zodResolver(formSchema),
	})

	const createNetwork = useMutation({
		...postNetworksFabricByIdNodesMutation({
			path: {
				id: networkId,
			},
		}),
		onSuccess: () => {
			toast.success('Node added successfully')
			onNodeAdded()
			setOpen(false)
			form.reset()
		},
		onError: () => {
			toast.error('Failed to add node')
		},
	})

	const onSubmit = async (data: FormValues) => {
		try {
			setIsLoading(true)
			const selectedNode = availableNodes.find(node => node.id.toString() === data.nodeId)
			if (!selectedNode) return

			const role = selectedNode.nodeType.toLowerCase().includes('orderer') ? 'orderer' : 'peer'
			
			createNetwork.mutate({
				path: {
					id: networkId,
				},
				body: {
					nodeId: parseInt(data.nodeId),
					role,
				},
			})
		} finally {
			setIsLoading(false)
		}
	}

	const selectedNode = availableNodes.find(
		node => node.id.toString() === form.watch('nodeId')
	)
	const nodeRole = selectedNode?.nodeType.toLowerCase().includes('orderer') ? 'orderer' : 'peer'

	return (
		<Dialog open={open} onOpenChange={setOpen}>
			<DialogTrigger asChild>
				<Button size="sm">
					<Plus className="mr-2 h-4 w-4" />
					Add Node
				</Button>
			</DialogTrigger>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>Add Node to Network</DialogTitle>
					<DialogDescription>Select a node to add to this network.</DialogDescription>
				</DialogHeader>
				<Form {...form}>
					<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
						<FormField
							control={form.control}
							name="nodeId"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Node</FormLabel>
									<Select 
										onValueChange={field.onChange}
										defaultValue={field.value}
									>
										<FormControl>
											<SelectTrigger>
												<SelectValue placeholder="Select a node" />
											</SelectTrigger>
										</FormControl>
										<SelectContent>
											{availableNodes.length === 0 ? (
												<SelectItem value="empty" disabled>
													No available nodes
												</SelectItem>
											) : (
												availableNodes.map((node) => (
													<SelectItem key={node.id} value={node.id.toString()}>
														{node.name} ({node.nodeType})
													</SelectItem>
												))
											)}
										</SelectContent>
									</Select>
									<FormMessage />
								</FormItem>
							)}
						/>
						{selectedNode && (
							<div className="space-y-1.5">
								<div className="text-sm font-medium">Role</div>
								<div className="text-sm text-muted-foreground">
									{nodeRole === 'orderer' ? 'Orderer Node' : 'Peer Node'}
								</div>
							</div>
						)}
						<div className="flex justify-end space-x-2">
							<Button variant="outline" type="button" onClick={() => setOpen(false)}>
								Cancel
							</Button>
							<Button type="submit" disabled={isLoading || availableNodes.length === 0}>
								{isLoading ? 'Adding...' : 'Add Node'}
							</Button>
						</div>
					</form>
				</Form>
			</DialogContent>
		</Dialog>
	)
}
