import { getChaincodeProjectsByIdOptions } from '@/api/client/@tanstack/react-query.gen'
import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

export default function ChaincodeProjectDetailPage() {
	const { id } = useParams<{ id: string }>()

	// Fetch project details
	const { data: project, isLoading } = useQuery({
		...getChaincodeProjectsByIdOptions({
			path: {
				id: Number(id),
			},
		}),
	})

	// Log project details for now
	console.log('Project details:', project)

	if (isLoading) {
		return (
			<div className="flex-1 p-8">
				<div className="text-center text-muted-foreground">Loading project details...</div>
			</div>
		)
	}

	if (!project) {
		return (
			<div className="flex-1 p-8">
				<div className="text-center text-red-500">Project not found</div>
			</div>
		)
	}

	return (
		<div className="flex-1 p-8">
			<div className="mb-6">
				<h1 className="text-2xl font-semibold">{project.name}</h1>
				<p className="text-muted-foreground">{project.description}</p>
			</div>

			{/* Placeholder for project details */}
			<div className="grid gap-6">
				<div className="rounded-lg border p-6">
					<h2 className="text-lg font-semibold mb-4">Project Information</h2>
					<div className="space-y-2 text-sm">
						<div>
							<span className="font-medium">ID:</span> {project.id}
						</div>
						<div>
							<span className="font-medium">Network ID:</span> {project.networkId}
						</div>
						<div>
							<span className="font-medium">Boilerplate:</span> {project.boilerplate}
						</div>
						<div>
							<span className="font-medium">Status:</span> {project.status}
						</div>
						{project.containerPort && (
							<div>
								<span className="font-medium">Container Port:</span> {project.containerPort}
							</div>
						)}
						{project.lastStartedAt && (
							<div>
								<span className="font-medium">Last Started:</span> {new Date(project.lastStartedAt).toLocaleString()}
							</div>
						)}
						{project.lastStoppedAt && (
							<div>
								<span className="font-medium">Last Stopped:</span> {new Date(project.lastStoppedAt).toLocaleString()}
							</div>
						)}
					</div>
				</div>
			</div>
		</div>
	)
} 