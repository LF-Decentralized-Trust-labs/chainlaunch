import { getChaincodeProjectsByIdOptions, putChaincodeProjectsByIdEndorsementPolicyMutation } from '@/api/client/@tanstack/react-query.gen'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query'
import { useParams, useNavigate } from 'react-router-dom'
import { Code } from 'lucide-react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogTrigger } from '@/components/ui/dialog'
import { Form, FormField, FormItem, FormLabel, FormControl, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { useForm } from 'react-hook-form'
import { useState } from 'react'
import { toast } from 'sonner'
import * as z from 'zod'
import { zodResolver } from '@hookform/resolvers/zod'

const endorsementPolicySchema = z.object({
	endorsementPolicy: z.string().min(1, 'Endorsement policy is required'),
})

type EndorsementPolicyFormValues = z.infer<typeof endorsementPolicySchema>

export default function ChaincodeProjectDetailPage() {
	const { id } = useParams()
	const navigate = useNavigate()
	const projectId = parseInt(id || '0', 10)

	const { data: project, isLoading, error, refetch } = useQuery({
		...getChaincodeProjectsByIdOptions({ path: { id: projectId } }),
		enabled: !!projectId,
	})

	const [isDialogOpen, setIsDialogOpen] = useState(false)
	const form = useForm<EndorsementPolicyFormValues>({
		resolver: zodResolver(endorsementPolicySchema),
		defaultValues: {
			endorsementPolicy: project?.endorsementPolicy || '',
		},
	})

	const [updating, setUpdating] = useState(false)
	const queryClient = useQueryClient()
	const updateEndorsementPolicyMutation = useMutation(putChaincodeProjectsByIdEndorsementPolicyMutation())

	const handleUpdate = async (data: EndorsementPolicyFormValues) => {
		setUpdating(true)
		try {
			await updateEndorsementPolicyMutation.mutateAsync({
				path: { id: projectId },
				body: { endorsementPolicy: data.endorsementPolicy },
			})
			toast.success('Endorsement policy updated')
			setIsDialogOpen(false)
			await queryClient.invalidateQueries({ queryKey: ['getChaincodeProjectsById', { path: { id: projectId } }] })
			await refetch()
		} catch (err: any) {
			toast.error('Failed to update endorsement policy', { description: err?.message })
		} finally {
			setUpdating(false)
		}
	}

	if (isLoading) return <div className="container p-8">Loading...</div>
	if (error) return <div className="container p-8 text-red-500">Error loading project</div>
	if (!project) return <div className="container p-8">Project not found</div>

	return (
		<div className="container p-8">
			<div className="flex justify-between items-center mb-6">
				<h1 className="text-2xl font-bold">{project.name}</h1>
				<div className="flex gap-2">
					<Button onClick={() => navigate(`/sc/fabric/projects/chaincodes/${project.id}/editor`)}>
						<Code className="mr-2 h-4 w-4" />
						Open Editor
					</Button>
					<Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
						<DialogTrigger asChild>
							<Button variant="outline">Update Endorsement Policy</Button>
						</DialogTrigger>
						<DialogContent>
							<DialogHeader>
								<DialogTitle>Update Endorsement Policy</DialogTitle>
							</DialogHeader>
							<Form {...form}>
								<form onSubmit={form.handleSubmit(handleUpdate)} className="space-y-4">
									<FormField
										control={form.control}
										name="endorsementPolicy"
										render={({ field }) => (
											<FormItem>
												<FormLabel>Endorsement Policy</FormLabel>
												<FormControl>
													<Input placeholder="e.g. OR('Org1MSP.member')" {...field} />
												</FormControl>
												<FormMessage />
												<div className="text-xs text-muted-foreground mt-1">
													Example policies:
													<ul className="list-disc list-inside mt-1">
														<li>OR('Org1MSP.member') - Any member of Org1</li>
														<li>AND('Org1MSP.member', 'Org2MSP.member') - Both Org1 and Org2 members</li>
														<li>OR('Org1MSP.member', 'Org2MSP.member') - Any member of Org1 or Org2</li>
													</ul>
												</div>
											</FormItem>
										)}
									/>
									<DialogFooter>
										<Button type="submit" disabled={updating}>
											{updating ? 'Updating...' : 'Update'}
										</Button>
									</DialogFooter>
								</form>
							</Form>
						</DialogContent>
					</Dialog>
				</div>
			</div>

			<div className="grid gap-4">
				<Card>
					<CardHeader>
						<CardTitle>Project Details</CardTitle>
						<CardDescription>Information about this chaincode project</CardDescription>
					</CardHeader>
					<CardContent>
						<div className="grid gap-4">
							<div>
								<h3 className="font-semibold mb-1">Description</h3>
								<p className="text-muted-foreground">{project.description || 'No description provided'}</p>
							</div>
							<div>
								<h3 className="font-semibold mb-1">ID</h3>
								<p className="text-muted-foreground">{project.id}</p>
							</div>
							<div>
								<h3 className="font-semibold mb-1">Network ID</h3>
								<p className="text-muted-foreground">{project.networkId}</p>
							</div>
							<div>
								<h3 className="font-semibold mb-1">Boilerplate</h3>
								<p className="text-muted-foreground">{project.boilerplate}</p>
							</div>
							<div>
								<h3 className="font-semibold mb-1">Status</h3>
								<p className="text-muted-foreground">{project.status}</p>
							</div>
							{project.endorsementPolicy && (
								<div>
									<h3 className="font-semibold mb-1">Endorsement Policy</h3>
									<p className="text-muted-foreground">{project.endorsementPolicy}</p>
								</div>
							)}
							{project.containerPort && (
								<div>
									<h3 className="font-semibold mb-1">Container Port</h3>
									<p className="text-muted-foreground">{project.containerPort}</p>
								</div>
							)}
							{project.lastStartedAt && (
								<div>
									<h3 className="font-semibold mb-1">Last Started</h3>
									<p className="text-muted-foreground">{new Date(project.lastStartedAt).toLocaleString()}</p>
								</div>
							)}
							{project.lastStoppedAt && (
								<div>
									<h3 className="font-semibold mb-1">Last Stopped</h3>
									<p className="text-muted-foreground">{new Date(project.lastStoppedAt).toLocaleString()}</p>
								</div>
							)}
						</div>
					</CardContent>
				</Card>
			</div>
		</div>
	)
} 