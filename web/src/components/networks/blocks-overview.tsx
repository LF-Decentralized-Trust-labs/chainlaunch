import { getNetworksFabricByIdBlocksOptions, getNetworksFabricByIdInfoOptions, getNetworksFabricByIdOptions } from '@/api/client/@tanstack/react-query.gen'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { useQuery } from '@tanstack/react-query'
import { formatDistanceToNow } from 'date-fns'
import { Link, useParams } from 'react-router-dom'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { useMemo } from 'react'
import { decodeBlockToJson } from '@/utils/block'

export function BlocksOverview() {
	const { id } = useParams<{ id: string }>()
	const networkId = parseInt(id || '0')

	const { data: chainInfo, isLoading: chainLoading } = useQuery({
		...getNetworksFabricByIdInfoOptions({
			path: { id: networkId },
		}),
	})
	const { data: networkInfo, isLoading: networkLoading } = useQuery({
		...getNetworksFabricByIdOptions({
			path: { id: networkId },
		}),
	})

	const { data: blocksResponse, isLoading: blocksLoading } = useQuery({
		...getNetworksFabricByIdBlocksOptions({
			path: { id: networkId },
			query: {
				limit: 10,
				offset: 0,
				reverse: true,
			},
		}),
	})
	const blocks = useMemo(() => blocksResponse?.blocks?.sort((a, b) => b.number! - a.number!) || [], [blocksResponse?.blocks])
	const lastBlock = useMemo(() => blocks[0], [blocks])
	const lastBlockJson = useMemo(() => (lastBlock ? decodeBlockToJson(lastBlock.data as unknown as string) : null), [lastBlock])
	console.log(lastBlockJson)
	if (chainLoading || networkLoading || blocksLoading) {
		return (
			<div className="space-y-6 p-4">
				<Skeleton className="h-[200px] w-full" />
				<Skeleton className="h-[400px] w-full" />
			</div>
		)
	}

	return (
		<div className="space-y-6 p-4	">
			<div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
				<Card>
					<CardHeader>
						<CardTitle>Latest Block</CardTitle>
					</CardHeader>
					<CardContent>
						<p className="text-2xl font-bold">#{chainInfo?.height}</p>
						<p className="text-sm text-muted-foreground">{/* {chainInfo && formatDistanceToNow(new Date(chainInfo.timestamp || ''), { addSuffix: true })} */}</p>
					</CardContent>
				</Card>
				<Card>
					<CardHeader>
						<CardTitle>Chain ID</CardTitle>
					</CardHeader>
					<CardContent>
						<p className="text-2xl font-bold">{networkInfo?.name}</p>
						<p className="text-sm text-muted-foreground">Network Identifier</p>
					</CardContent>
				</Card>
				<Card>
					<CardHeader>
						<CardTitle>Total Transactions</CardTitle>
					</CardHeader>
					<CardContent>
						<p className="text-2xl font-bold">{lastBlockJson?.data?.data?.length || 0}</p>
						<p className="text-sm text-muted-foreground">In Latest Block</p>
					</CardContent>
				</Card>
			</div>

			<Card>
				<CardHeader>
					<CardTitle>Recent Blocks</CardTitle>
				</CardHeader>
				<CardContent>
					<Table>
						<TableHeader>
							<TableRow>
								<TableHead>Block</TableHead>
								<TableHead>Hash</TableHead>
								<TableHead>Time</TableHead>
								<TableHead>Transactions</TableHead>
								<TableHead className="text-right">Actions</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{blocks.map((block) => (
								<TableRow key={block.number}>
									<TableCell className="font-medium">#{block.number}</TableCell>
									<TableCell className="font-mono text-sm">
										{block.hash && (
											<>
												{block.hash.substring(0, 10)}...{block.hash.substring(block.hash.length - 10)}
											</>
										)}
									</TableCell>
									<TableCell>{formatDistanceToNow(new Date(block.timestamp || ''), { addSuffix: true })}</TableCell>
									<TableCell>{block.tx_count}</TableCell>
									<TableCell className="text-right">
										<Button variant="ghost" size="sm" asChild>
											<Link to={`/networks/${id}/blocks/${block.number}`}>View Details</Link>
										</Button>
									</TableCell>
								</TableRow>
							))}
						</TableBody>
					</Table>
				</CardContent>
			</Card>
		</div>
	)
}
