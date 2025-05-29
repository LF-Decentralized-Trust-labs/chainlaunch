import { getNetworksFabricOptions, getScFabricChaincodesOptions } from '@/api/client/@tanstack/react-query.gen'
import { getScFabricChaincodes, postScFabricChaincodes } from '@/api/client/sdk.gen'
import type { ChainlaunchdeployFabricChaincodeDetail } from '@/api/client/types.gen'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { AlertCircle, CheckCircle2, Clock, Plus } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import * as z from 'zod'

const chaincodeFormSchema = z.object({
	name: z.string().min(1, 'Name is required'),
	networkId: z.string().min(1, 'Network is required'),
})

type ChaincodeFormValues = z.infer<typeof chaincodeFormSchema>

export default function FabricChaincodesPage() {
	const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false)
	const [createError, setCreateError] = useState<string | null>(null)
	const queryClient = useQueryClient()
	const form = useForm<ChaincodeFormValues>({
		resolver: zodResolver(chaincodeFormSchema),
		defaultValues: { name: '', networkId: '' },
	})
	const navigate = useNavigate()

	// Fetch networks
	const { data: networks } = useQuery({
		...getNetworksFabricOptions({
			query: {
				limit: 10,
				offset: 0,
			},
		}),
	})

	// Fetch chaincodes
	const { data: chaincodesResponse } = useQuery({
		...getScFabricChaincodesOptions(),
	})

	// Create chaincode mutation
	const createChaincodeMutation = useMutation({
		mutationFn: async (data: ChaincodeFormValues) => {
			const response = await postScFabricChaincodes({
				body: {
					name: data.name,
					network_id: Number(data.networkId),
				},
			})
			return response.data
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['fabricChaincodes'] })
			setIsCreateDialogOpen(false)
			form.reset()
			setCreateError(null)
		},
		onError: (error: any) => {
			let message = 'Failed to create chaincode.'
			if (error?.response?.data?.message) {
				message = error.response.data.message
			} else if (error?.message) {
				message = error.message
			}
			setCreateError(message)
		},
	})

	const onSubmit = async (data: ChaincodeFormValues) => {
		setCreateError(null)
		await createChaincodeMutation.mutateAsync(data)
	}

	return (
		<div className="flex-1 p-8">
			<div className="mb-6">
				<div className="flex items-center justify-between">
					<div>
						<h1 className="text-2xl font-semibold">Chaincodes</h1>
						<p className="text-muted-foreground">Manage chaincodes for your Fabric networks</p>
					</div>
					<Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
						<DialogTrigger asChild>
							<Button>
								<Plus className="mr-2 h-4 w-4" />
								Create Chaincode
							</Button>
						</DialogTrigger>
						<DialogContent>
							<DialogHeader>
								<DialogTitle>Create Chaincode</DialogTitle>
								<DialogDescription>Define a new chaincode for your Fabric network.</DialogDescription>
							</DialogHeader>
							<Form {...form}>
								<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
									<FormField
										control={form.control}
										name="name"
										render={({ field }) => (
											<FormItem>
												<FormLabel>Name</FormLabel>
												<FormControl>
													<Input {...field} />
												</FormControl>
												<FormMessage />
											</FormItem>
										)}
									/>
									<FormField
										control={form.control}
										name="networkId"
										render={({ field }) => (
											<FormItem>
												<FormLabel>Network</FormLabel>
												<FormControl>
													<select {...field} className="w-full border rounded p-2">
														<option value="">Select a network</option>
														{networks?.networks?.map((n) => (
															<option key={n.id} value={n.id}>
																{n.name}
															</option>
														))}
													</select>
												</FormControl>
												<FormMessage />
											</FormItem>
										)}
									/>
									{createError && (
										<div className="rounded bg-red-100 border border-red-300 text-red-700 px-3 py-2 text-sm mb-2" role="alert">
											{createError}
										</div>
									)}
									<DialogFooter>
										<Button type="submit" disabled={createChaincodeMutation.isPending}>
											{createChaincodeMutation.isPending ? 'Creating...' : 'Create'}
										</Button>
									</DialogFooter>
								</form>
							</Form>
						</DialogContent>
					</Dialog>
				</div>
			</div>
			<div className="space-y-6">
				{!chaincodesResponse?.chaincodes?.length ? (
					<Card className="p-6 text-center text-muted-foreground">No chaincodes yet.</Card>
				) : (
					chaincodesResponse.chaincodes.map((detail) => (
						<Card key={detail.id} className="p-6 flex items-center justify-between">
							<div>
								<div className="font-semibold text-lg">{detail.name}</div>
								<div className="text-sm text-muted-foreground">Network: {networks?.networks?.find((n) => n.id === detail.network_id)?.name}</div>
							</div>
							<Button
								variant="outline"
								size="sm"
								onClick={() =>
									navigate(`/sc/fabric/chaincodes/${detail.id}`)
								}
							>
								View Details
							</Button>
						</Card>
					))
				)}
			</div>
		</div>
	)
}
