import { getChaincodeProjectsByIdOptions } from '@/api/client/@tanstack/react-query.gen'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useQuery } from '@tanstack/react-query'
import { useParams, useNavigate } from 'react-router-dom'
import { Code } from 'lucide-react'

export default function ChaincodeProjectDetailPage() {
	const { id } = useParams()
	const navigate = useNavigate()
	const projectId = parseInt(id || '0', 10)

	const { data: project, isLoading, error } = useQuery({
		...getChaincodeProjectsByIdOptions({ path: { id: projectId } }),
		enabled: !!projectId,
	})

	if (isLoading) return <div className="container p-8">Loading...</div>
	if (error) return <div className="container p-8 text-red-500">Error loading project</div>
	if (!project) return <div className="container p-8">Project not found</div>

	return (
		<div className="container p-8">
			<div className="flex justify-between items-center mb-6">
				<h1 className="text-2xl font-bold">{project.name}</h1>
				<Button onClick={() => navigate(`/sc/fabric/projects/chaincodes/${project.id}/editor`)}>
					<Code className="mr-2 h-4 w-4" />
					Open Editor
				</Button>
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