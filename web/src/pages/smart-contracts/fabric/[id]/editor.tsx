import { getChaincodeProjectsByIdOptions, postChaincodeProjectsByIdStartMutation, postChaincodeProjectsByIdStopMutation } from '@/api/client/@tanstack/react-query.gen'
import { CodeEditor } from '@/components/editor/CodeEditor'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { useMutation, useQuery } from '@tanstack/react-query'
import { ArrowLeft, Code2, PlayCircle } from 'lucide-react'
import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { toast } from 'sonner'

export default function ChaincodeProjectEditorPage() {
	const { id } = useParams()
	const navigate = useNavigate()
	const projectId = parseInt(id || '0', 10)

	const { data: project, isLoading, error, refetch } = useQuery(getChaincodeProjectsByIdOptions({ path: { id: projectId } }))

	const startMutation = useMutation(postChaincodeProjectsByIdStartMutation())
	const stopMutation = useMutation(postChaincodeProjectsByIdStopMutation())

	const [mode, setMode] = useState<'editor' | 'playground'>('editor')

	const handleStart = async () => {
		try {
			const toastId = toast.loading('Starting project...')
			await startMutation.mutateAsync({ path: { id: projectId } })
			toast.success('Project started!', { id: toastId })
			await refetch()
		} catch (err: any) {
			toast.error('Failed to start project', { description: err?.message })
		}
	}

	const handleStop = async () => {
		try {
			const toastId = toast.loading('Stopping project...')
			await stopMutation.mutateAsync({ path: { id: projectId } })
			toast.success('Project stopped!', { id: toastId })
			await refetch()
		} catch (err: any) {
			toast.error('Failed to stop project', { description: err?.message })
		}
	}

	if (isLoading) {
		return (
			<div className="h-screen flex flex-col">
				<div className="flex items-center gap-4 p-4 border-b bg-background">
					<Skeleton className="h-10 w-10 rounded-full" />
					<Skeleton className="h-6 w-40 rounded" />
					<div className="ml-auto flex gap-2 items-center">
						<Skeleton className="h-10 w-10 rounded" />
						<Skeleton className="h-10 w-10 rounded" />
						<Skeleton className="h-10 w-28 rounded" />
						<Skeleton className="h-4 w-20 rounded" />
					</div>
				</div>
				<div className="flex-1 p-4">
					<Skeleton className="h-full w-full rounded" />
				</div>
			</div>
		)
	}

	if (error) {
		return (
			<div className="h-screen flex flex-col items-center justify-center">
				<Alert variant="destructive" className="max-w-md w-full">
					<AlertTitle>Error loading project</AlertTitle>
					<AlertDescription>
						{error instanceof Error ? error.message : 'An unknown error occurred.'}
					</AlertDescription>
					<Button variant="outline" onClick={() => refetch()} className="mt-4">Retry</Button>
				</Alert>
			</div>
		)
	}

	return (
		<div className="h-screen flex flex-col">
			<div className="flex items-center gap-4 p-4 border-b bg-background">
				<Button variant="ghost" size="icon" onClick={() => navigate(`/sc/fabric/projects/chaincodes/${projectId}`)}>
					<ArrowLeft className="h-4 w-4" />
				</Button>
				<h1 className="text-lg font-semibold">Chaincode Editor</h1>
				<div className="ml-auto flex gap-2 items-center">
					<Button variant={mode === 'editor' ? 'default' : 'outline'} size="icon" onClick={() => setMode('editor')} title="Editor mode">
						<Code2 className="h-5 w-5" />
					</Button>
					<Button variant={mode === 'playground' ? 'default' : 'outline'} size="icon" onClick={() => setMode('playground')} title="Playground mode">
						<PlayCircle className="h-5 w-5" />
					</Button>
					{project?.status === 'running' ? (
						<Button onClick={handleStop} disabled={stopMutation.isPending} variant="destructive">
							{stopMutation.isPending ? 'Stopping...' : 'Stop Project'}
						</Button>
					) : (
						<Button onClick={handleStart} disabled={startMutation.isPending}>
							{startMutation.isPending ? 'Starting...' : 'Start Project'}
						</Button>
					)}
					<span className="text-muted-foreground text-xs">Status: {project?.status}</span>
				</div>
			</div>
			<div className="flex-1 p-4">
				<CodeEditor mode={mode} projectId={projectId} key={mode} chaincodeProject={project} />
			</div>
		</div>
	)
}
