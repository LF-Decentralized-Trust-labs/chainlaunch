import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { PlayCircle, Search, RotateCcw } from 'lucide-react'
import { ScrollArea } from '@/components/ui/scroll-area'
import { FabricKeySelect } from '@/components/FabricKeySelect'
import { postChaincodeProjectsByIdInvoke, postChaincodeProjectsByIdQuery } from '@/api/client'
import { toast } from 'sonner'
import { useMemo, useState, useEffect, useCallback, useRef } from 'react'

interface PlaygroundProps {
	projectId: number
	networkId: number
}

interface Operation {
	fn: string
	args: string
	selectedKey: { orgId: number; keyId: number } | undefined
	type: 'invoke' | 'query'
	timestamp: number
}

function isValidResponse(r: any): r is { type: 'invoke' | 'query'; response: any; error?: string; timestamp: number } {
	return r && (r.type === 'invoke' || r.type === 'query') && typeof r.timestamp === 'number'
}
function renderResponseContent(content: any) {
	if (!content) {
		return <span className="mt-1 italic">Empty response</span>
	}

	if (typeof content === 'string') {
		try {
			const parsed = JSON.parse(content)
			return <pre className="whitespace-pre-wrap break-all text-sm mt-1">{JSON.stringify(parsed, null, 2)}</pre>
		} catch {
			return <pre className="whitespace-pre-wrap break-all text-sm mt-1">{content}</pre>
		}
	} else if (typeof content === 'object' && content !== null) {
		return <pre className="whitespace-pre-wrap break-all text-sm mt-1">{JSON.stringify(content, null, 2)}</pre>
	} else {
		return <pre className="whitespace-pre-wrap break-all text-sm mt-1">{String(content)}</pre>
	}
}

export function Playground({ projectId, networkId }: PlaygroundProps) {
	const [fn, setFn] = useState('')
	const [args, setArgs] = useState('')
	const [selectedKey, setSelectedKey] = useState<{ orgId: number; keyId: number } | undefined>(undefined)
	const [responses, setResponses] = useState<{ type: 'invoke' | 'query'; response: any; error?: string; timestamp: number; fn: string; args: string }[]>([])
	const [operations, setOperations] = useState<Operation[]>([])
	const [loadingInvoke, setLoadingInvoke] = useState(false)
	const [loadingQuery, setLoadingQuery] = useState(false)
	const didMount = useRef(false)

	// Load state from localStorage on mount
	useEffect(() => {
		const saved = localStorage.getItem(`playground-state-${projectId}`)
		if (saved) {
			try {
				const parsed = JSON.parse(saved)
				setFn(parsed.fn || '')
				setArgs(parsed.args || '')
				setSelectedKey(parsed.selectedKey)
				setOperations(parsed.operations || [])
				setResponses(Array.isArray(parsed.responses) ? (parsed.responses.filter(isValidResponse) as { type: 'invoke' | 'query'; response: any; error?: string; timestamp: number; fn: string; args: string }[]) : [])
			} catch {}
		}
	}, [projectId])

	const saveToLocalStorage = (
		nextState: Partial<{
			fn: string
			args: string
			selectedKey: { orgId: number; keyId: number } | undefined
			operations: Operation[]
			responses: any[]
		}> = {}
	) => {
		localStorage.setItem(
			`playground-state-${projectId}`,
			JSON.stringify({
				fn,
				args,
				selectedKey,
				operations,
				responses,
				...nextState,
			})
		)
	}

	const saveOperation = useCallback(
		(type: 'invoke' | 'query', nextResponses?: any[]) => {
			if (!fn || !selectedKey) return
			const op: Operation = {
				fn,
				args,
				selectedKey,
				type,
				timestamp: Date.now(),
			}
			setOperations((prev) => {
				const next = [
					op,
					...prev.filter((o) => !(o.fn === fn && o.args === args && o.type === type && o.selectedKey?.orgId === selectedKey.orgId && o.selectedKey?.keyId === selectedKey.keyId)),
				]
				const sliced = next.slice(0, 10)
				saveToLocalStorage({ operations: sliced, responses: nextResponses ?? responses })
				return sliced
			})
		},
		[fn, args, selectedKey, responses]
	)

	const handleInvoke = async () => {
		if (!selectedKey) return
		setLoadingInvoke(true)
		const toastId = toast.loading('Invoking...')
		try {
			const res = await postChaincodeProjectsByIdInvoke({
				path: { id: projectId },
				body: {
					function: fn,
					args: args
						.split(',')
						.map((a) => a.trim())
						.filter(Boolean),
					keyId: selectedKey.keyId,
					orgId: selectedKey.orgId,
				},
			})
			toast.dismiss(toastId)
			let nextResponses
			if (res.error) {
				nextResponses = [...responses, { type: 'invoke', response: null, error: res.error.message, timestamp: Date.now(), fn, args }]
				setResponses(nextResponses as { type: 'invoke' | 'query'; response: any; error?: string; timestamp: number; fn: string; args: string }[])
			} else {
				nextResponses = [...responses, { type: 'invoke', response: res.data, timestamp: Date.now(), fn, args }]
				setResponses(nextResponses as { type: 'invoke' | 'query'; response: any; error?: string; timestamp: number; fn: string; args: string }[])
			}
			saveOperation('invoke', nextResponses)
		} catch (err: any) {
			toast.dismiss(toastId)
			const msg = err?.response?.data?.message || err?.message || 'Unknown error'
			const nextResponses = [...responses, { type: 'invoke', response: null, error: msg, timestamp: Date.now(), fn, args }]
			setResponses(nextResponses as { type: 'invoke' | 'query'; response: any; error?: string; timestamp: number; fn: string; args: string }[])
			saveOperation('invoke', nextResponses)
		} finally {
			setLoadingInvoke(false)
		}
	}

	const handleQuery = async () => {
		if (!selectedKey) return
		setLoadingQuery(true)
		const toastId = toast.loading('Querying...')
		try {
			const res = await postChaincodeProjectsByIdQuery({
				path: { id: projectId },
				body: {
					function: fn,
					args: args
						.split(',')
						.map((a) => a.trim())
						.filter(Boolean),
					keyId: selectedKey.keyId,
					orgId: selectedKey.orgId,
				},
			})
			toast.dismiss(toastId)
			let nextResponses
			if (res.error) {
				nextResponses = [...responses, { type: 'query', response: null, error: res.error.message, timestamp: Date.now(), fn, args }]
				setResponses(nextResponses as { type: 'invoke' | 'query'; response: any; error?: string; timestamp: number; fn: string; args: string }[])
			} else {
				nextResponses = [...responses, { type: 'query', response: res.data, timestamp: Date.now(), fn, args }]
				setResponses(nextResponses as { type: 'invoke' | 'query'; response: any; error?: string; timestamp: number; fn: string; args: string }[])
			}
			saveOperation('query', nextResponses)
		} catch (err: any) {
			toast.dismiss(toastId)
			const msg = err?.response?.data?.message || err?.message || 'Unknown error'
			const nextResponses = [...responses, { type: 'query', response: null, error: msg, timestamp: Date.now(), fn, args }]
			setResponses(nextResponses as { type: 'invoke' | 'query'; response: any; error?: string; timestamp: number; fn: string; args: string }[])
			saveOperation('query', nextResponses)
		} finally {
			setLoadingQuery(false)
		}
	}

	const restoreAndRun = (op: Operation) => {
		setFn(op.fn)
		setArgs(op.args)
		setSelectedKey(op.selectedKey)
		saveToLocalStorage({ fn: op.fn, args: op.args, selectedKey: op.selectedKey })
		if (op.type === 'invoke') {
			handleInvoke()
		} else {
			handleQuery()
		}
	}

	const restoreOnly = (op: Operation) => {
		setFn(op.fn)
		setArgs(op.args)
		setSelectedKey(op.selectedKey)
		saveToLocalStorage({ fn: op.fn, args: op.args, selectedKey: op.selectedKey })
	}

	const sortedResponses = useMemo(() => responses.sort((a, b) => b.timestamp - a.timestamp), [responses])

	// Persist fn and args only after initial mount and if both have value
	useEffect(() => {
		if (didMount.current) {
			if (fn && args) {
				saveToLocalStorage({ fn, args })
			}
		} else {
			didMount.current = true
		}
	}, [fn, args])

	return (
		<div className="grid grid-cols-1 md:grid-cols-6 gap-8">
			{/* Playground form (left) */}
			<ScrollArea className="md:col-span-2 min-w-[320px] max-w-[480px] border rounded bg-background shadow-sm my-4">
				<div className="px-4 py-4 grid gap-4 h-full">
					<h2 className="text-xl font-bold mb-2 grid grid-flow-col auto-cols-max items-center gap-2">
						<PlayCircle className="h-5 w-5" /> Playground
					</h2>
					<Label>Key & Organization</Label>
					<FabricKeySelect value={selectedKey} onChange={setSelectedKey} />
					<Label htmlFor="fn">Function name</Label>
					<Input id="fn" value={fn} onChange={(e) => setFn(e.target.value)} placeholder="e.g. queryAsset" />
					<Label htmlFor="args">Arguments (comma separated)</Label>
					<Input id="args" value={args} onChange={(e) => setArgs(e.target.value)} placeholder="e.g. asset1, 100" />
					<div className="grid grid-flow-col auto-cols-max gap-2 mt-2">
						<Button onClick={handleInvoke} disabled={loadingInvoke || !fn || !selectedKey}>
							<PlayCircle className="h-4 w-4 mr-2" />
							Invoke
						</Button>
						<Button onClick={handleQuery} disabled={loadingQuery || !fn || !selectedKey} variant="secondary">
							<Search className="h-4 w-4 mr-2" />
							Query
						</Button>
					</div>
					{/* Recent Operations */}
					{operations.length > 0 && (
						<div className="mt-6">
							<h3 className="text-md font-semibold mb-2">Recent Operations</h3>
							<div className="grid gap-2">
								{operations.map((op) => (
									<div key={op.timestamp} className="grid grid-flow-col auto-cols-max items-center gap-2 p-2 border rounded bg-muted/50">
										<div className="flex-1">
											<div className="text-xs text-muted-foreground mb-1 grid grid-flow-col auto-cols-max items-center gap-1">
												{op.type === 'invoke' ? <PlayCircle className="h-4 w-4" /> : <Search className="h-4 w-4" />}
												{op.type.toUpperCase()} &middot; {new Date(op.timestamp).toLocaleTimeString()}
											</div>
											<div className="text-sm font-mono">
												fn: <span className="font-semibold">{op.fn}</span>
											</div>
											<div className="text-xs text-muted-foreground">args: {op.args}</div>
											<div className="text-xs text-muted-foreground">
												org: {op.selectedKey?.orgId}, key: {op.selectedKey?.keyId}
											</div>
										</div>
										<Button size="icon" variant="ghost" onClick={() => restoreOnly(op)} className="rounded-full" title="Restore">
											<RotateCcw className="h-4 w-4" />
										</Button>
										<Button
											size="icon"
											variant="secondary"
											onClick={() => {
												restoreOnly(op)
												handleInvoke()
											}}
											title="Restore & Invoke"
										>
											<PlayCircle className="h-4 w-4" />
										</Button>
										<Button
											size="icon"
											variant="secondary"
											onClick={() => {
												restoreOnly(op)
												handleQuery()
											}}
											title="Restore & Query"
										>
											<Search className="h-4 w-4" />
										</Button>
									</div>
								))}
							</div>
						</div>
					)}
				</div>
			</ScrollArea>
			{/* Responses (right) */}
			<ScrollArea className="md:col-span-4 max-h-[700px] border rounded bg-muted/30 p-2 my-4">
				{responses.length === 0 ? (
					<div className="text-muted-foreground text-center py-8">No responses yet</div>
				) : (
					<div className="grid gap-2">
						{sortedResponses.map((response) => (
							<div key={response.timestamp} className="p-3 border-b border-muted bg-background rounded">
								<div className="grid grid-flow-col auto-cols-max items-center gap-2 mb-1">
									<span className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{response.type === 'invoke' ? 'Invoke' : 'Query'}</span>
									<span className="text-xs text-muted-foreground">{new Date(response.timestamp).toLocaleTimeString()}</span>
								</div>
								<div className="text-xs text-muted-foreground mb-1">
									fn: <span className="font-semibold">{response.fn}</span> &nbsp; args: <span className="font-mono">{response.args}</span>
								</div>
								<div className="text-sm whitespace-pre-wrap break-all">
									{response.error ? (
										<span className="text-destructive">Error</span>
									) : (
										<>
											<div className="max-h-64 overflow-auto overflow-x-auto">{renderResponseContent(response.response.result)}</div>
											{networkId && response.response && !!response.response.blockNumber && !!response.response.transactionId && (
												<a
													href={`/networks/${networkId}/blocks/${response.response.blockNumber}`}
													className="inline-block mt-2 text-primary underline text-xs hover:text-primary/80"
													target="_blank"
													rel="noopener noreferrer"
												>
													View Block #{response.response.blockNumber}
												</a>
											)}
										</>
									)}
								</div>
							</div>
						))}
					</div>
				)}
			</ScrollArea>
		</div>
	)
}
