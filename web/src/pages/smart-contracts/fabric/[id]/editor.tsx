import { CodeEditor } from '@/components/editor/CodeEditor'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useParams, useNavigate } from 'react-router-dom'
import { ArrowLeft, Code2, PlayCircle } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getChaincodeProjectsByIdOptions, postChaincodeProjectsByIdStartMutation } from '@/api/client/@tanstack/react-query.gen'
import { postChaincodeProjectsByIdInvoke, postChaincodeProjectsByIdQuery } from '@/api/client'
import { toast } from 'sonner'
import { FabricKeySelect } from '@/components/FabricKeySelect'

export default function ChaincodeProjectEditorPage() {
	const { id } = useParams()
	const navigate = useNavigate()
	const projectId = parseInt(id || '0', 10)
	const queryClient = useQueryClient()

	const { data: project, isLoading, refetch } = useQuery(getChaincodeProjectsByIdOptions({ path: { id: projectId } }))

	const startMutation = useMutation({
		...postChaincodeProjectsByIdStartMutation(),
		onSuccess: async () => {
			await refetch()
			queryClient.invalidateQueries({ queryKey: ['getChaincodeProjectsByIdLogs', { path: { id: projectId } }] })
		},
	})

	useEffect(() => {
		if (!isLoading && project && project.status !== 'running') {
			toast.promise(
				startMutation.mutateAsync({ path: { id: projectId } }),
				{
					loading: 'Starting chaincode project...',
					success: 'Chaincode project started!',
					error: 'Failed to start chaincode project',
				}
			)
		}
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [isLoading, projectId, project?.status])

	const [mode, setMode] = useState<'editor' | 'playground'>('editor')
	const [fn, setFn] = useState('')
	const [args, setArgs] = useState('')
	const [selectedKey, setSelectedKey] = useState<{ orgId: number; keyId: number } | undefined>(undefined)
	const [invokeResult, setInvokeResult] = useState<any>(null)
	const [queryResult, setQueryResult] = useState<any>(null)
	const [loadingInvoke, setLoadingInvoke] = useState(false)
	const [loadingQuery, setLoadingQuery] = useState(false)

	const handleInvoke = async () => {
		if (!selectedKey) return
		setLoadingInvoke(true)
		setInvokeResult(null)
		try {
			const res = await postChaincodeProjectsByIdInvoke({
				path: { id: projectId },
				body: { function: fn, args: args.split(',').map((a) => a.trim()).filter(Boolean), keyId: selectedKey.keyId, orgId: selectedKey.orgId },
			})
			setInvokeResult(res.data)
			toast.success('Invoke transaction sent!')
		} catch (err: any) {
			toast.error('Invoke failed', { description: err?.message })
		} finally {
			setLoadingInvoke(false)
		}
	}

	const handleQuery = async () => {
		if (!selectedKey) return
		setLoadingQuery(true)
		setQueryResult(null)
		try {
			const res = await postChaincodeProjectsByIdQuery({
				path: { id: projectId },
				body: { function: fn, args: args.split(',').map((a) => a.trim()).filter(Boolean), keyId: selectedKey.keyId, orgId: selectedKey.orgId },
			})
			setQueryResult(res.data)
			toast.success('Query executed!')
		} catch (err: any) {
			toast.error('Query failed', { description: err?.message })
		} finally {
			setLoadingQuery(false)
		}
	}

	return (
		<div className="h-screen flex flex-col">
			<div className="flex items-center gap-4 p-4 border-b bg-background">
				<Button variant="ghost" size="icon" onClick={() => navigate(`/sc/fabric/projects/chaincodes/${projectId}`)}>
					<ArrowLeft className="h-4 w-4" />
				</Button>
				<h1 className="text-lg font-semibold">Chaincode Editor</h1>
				<div className="ml-auto flex gap-2">
					<Button variant={mode === 'editor' ? 'default' : 'outline'} size="icon" onClick={() => setMode('editor')} title="Editor mode">
						<Code2 className="h-5 w-5" />
					</Button>
					<Button variant={mode === 'playground' ? 'default' : 'outline'} size="icon" onClick={() => setMode('playground')} title="Playground mode">
						<PlayCircle className="h-5 w-5" />
					</Button>
				</div>
			</div>
			<div className="flex-1 p-4">
				<div className="h-full rounded-lg border bg-background">
					{mode === 'editor' ? (
						<CodeEditor />
					) : (
						<div className="p-6 max-w-xl mx-auto flex flex-col gap-6">
							<h2 className="text-xl font-bold mb-2 flex items-center gap-2"><PlayCircle className="h-5 w-5" /> Playground</h2>
							<div className="flex flex-col gap-4">
								<Label>Key & Organization</Label>
								<FabricKeySelect
									value={selectedKey}
									onChange={setSelectedKey}
								/>
								<Label htmlFor="fn">Function name</Label>
								<Input id="fn" value={fn} onChange={e => setFn(e.target.value)} placeholder="e.g. queryAsset" />
								<Label htmlFor="args">Arguments (comma separated)</Label>
								<Input id="args" value={args} onChange={e => setArgs(e.target.value)} placeholder="e.g. asset1, 100" />
								<div className="flex gap-2 mt-2">
									<Button onClick={handleInvoke} disabled={loadingInvoke || !fn || !selectedKey}>
										Invoke
									</Button>
									<Button onClick={handleQuery} disabled={loadingQuery || !fn || !selectedKey} variant="secondary">
										Query
									</Button>
								</div>
								{invokeResult && (
									<div className="bg-muted rounded p-4 mt-2">
										<div className="font-semibold mb-1">Invoke Result</div>
										<div className="space-y-2">
											{invokeResult.txId && (
												<div>
													<span className="text-muted-foreground">Transaction ID:</span>
													<pre className="whitespace-pre-wrap break-all text-sm mt-1">{invokeResult.txId}</pre>
												</div>
											)}
											{invokeResult.status && (
												<div>
													<span className="text-muted-foreground">Status:</span>
													<pre className="whitespace-pre-wrap break-all text-sm mt-1">{invokeResult.status}</pre>
												</div>
											)}
											{invokeResult.message && (
												<div>
													<span className="text-muted-foreground">Message:</span>
													<pre className="whitespace-pre-wrap break-all text-sm mt-1">{invokeResult.message}</pre>
												</div>
											)}
											{invokeResult.payload && (
												<div>
													<span className="text-muted-foreground">Payload:</span>
													<pre className="whitespace-pre-wrap break-all text-sm mt-1">{JSON.stringify(invokeResult.payload, null, 2)}</pre>
												</div>
											)}
										</div>
									</div>
								)}
								{queryResult && (
									<div className="bg-muted rounded p-4 mt-2">
										<div className="font-semibold mb-1">Query Result</div>
										<div className="space-y-2">
											{queryResult.payload && (
												<div>
													<span className="text-muted-foreground">Payload:</span>
													<pre className="whitespace-pre-wrap break-all text-sm mt-1">{JSON.stringify(queryResult.payload, null, 2)}</pre>
												</div>
											)}
											{queryResult.message && (
												<div>
													<span className="text-muted-foreground">Message:</span>
													<pre className="whitespace-pre-wrap break-all text-sm mt-1">{queryResult.message}</pre>
												</div>
											)}
											{queryResult.status && (
												<div>
													<span className="text-muted-foreground">Status:</span>
													<pre className="whitespace-pre-wrap break-all text-sm mt-1">{queryResult.status}</pre>
												</div>
											)}
										</div>
									</div>
								)}
							</div>
						</div>
					)}
				</div>
			</div>
		</div>
	)
} 