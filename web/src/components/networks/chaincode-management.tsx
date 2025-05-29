import { HttpNetworkResponse } from '@/api/client'
import {
	getNodesByIdChannelsByChannelIdChaincodesOptions,
	postScFabricDeployMutation,
	postScFabricPeerByPeerIdChaincodeApproveMutation,
	postScFabricPeerByPeerIdChaincodeCommitMutation,
	postScFabricPeerByPeerIdChaincodeInstallMutation,
} from '@/api/client/@tanstack/react-query.gen'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { useMutation, useQuery } from '@tanstack/react-query'
import { AlertTriangle, Code, FileCode, Loader2, Plus } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Badge } from '../ui/badge'

interface ChaincodeManagementProps {
	network: HttpNetworkResponse
	peerId: number
	channelName: string
	organizationId: number
}

const chaincodeFormSchema = z.object({
	name: z.string().min(1, 'Name is required'),
})

type ChaincodeFormValues = z.infer<typeof chaincodeFormSchema>

export function ChaincodeManagement({ }: ChaincodeManagementProps) {
	const [isDeployDialogOpen, setIsDeployDialogOpen] = useState(false)
	const [selectedChaincode, setSelectedChaincode] = useState<{ name: string } | null>(null)

	const form = useForm<ChaincodeFormValues>({
		resolver: zodResolver(chaincodeFormSchema),
		defaultValues: {
			name: '',
		},
	})

	const onSubmit = async (data: ChaincodeFormValues) => {
		try {
			// For now, just log the chaincode name
			console.log('Chaincode name:', data.name)
			toast.success('Chaincode name recorded successfully')
			setIsDeployDialogOpen(false)
		} catch (error: any) {
			toast.error('Failed to record chaincode name', {
				description: error.message,
			})
		}
	}

	return (
		<div className="space-y-6">
			<div className="flex items-center justify-between">
				<div className="flex items-center gap-4">
					<div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center">
						<Code className="h-6 w-6 text-primary" />
					</div>
					<div>
						<h2 className="text-lg font-semibold">Chaincode Management</h2>
						<p className="text-sm text-muted-foreground">Record chaincode names for your network</p>
					</div>
				</div>
				<Dialog open={isDeployDialogOpen} onOpenChange={setIsDeployDialogOpen}>
					<DialogTrigger asChild>
						<Button>
							<Plus className="mr-2 h-4 w-4" />
							Record Chaincode
						</Button>
					</DialogTrigger>
					<DialogContent>
						<DialogHeader>
							<DialogTitle>Record New Chaincode</DialogTitle>
							<DialogDescription>Record a new chaincode name for your network</DialogDescription>
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
								<DialogFooter>
									<Button type="submit">
										Record Chaincode
									</Button>
								</DialogFooter>
							</form>
						</Form>
					</DialogContent>
				</Dialog>
			</div>

			<Card className="p-6">
				<div className="flex items-center gap-4">
					<AlertTriangle className="h-5 w-5 text-muted-foreground" />
					<p className="text-sm text-muted-foreground">Chaincode recording is in development. Only names are being recorded for now.</p>
				</div>
			</Card>
		</div>
	)
} 