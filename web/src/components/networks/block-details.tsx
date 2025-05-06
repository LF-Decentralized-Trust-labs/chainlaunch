import { getNetworksFabricByIdBlocksByBlockNumOptions } from '@/api/client/@tanstack/react-query.gen'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@radix-ui/react-dropdown-menu'
import { useQuery } from '@tanstack/react-query'
import { ArrowLeft, EllipsisVertical } from 'lucide-react'
import { useMemo } from 'react'
import { Link, useParams } from 'react-router-dom'
import { decodeBlockToJson } from '@/utils/block'

export function BlockDetails() {
	const { id, blockNumber } = useParams<{ id: string; blockNumber: string }>()
	const networkId = parseInt(id || '0')
	const blockNum = parseInt(blockNumber || '0')

	const { data: blockResponse, isLoading } = useQuery({
		...getNetworksFabricByIdBlocksByBlockNumOptions({
			path: { id: networkId, blockNum },
		}),
	})
	const decodedBlock = useMemo(() => {
		if (!blockResponse?.block?.data) return null
		return decodeBlockToJson(blockResponse.block.data as unknown as string)
	}, [blockResponse?.block?.data])

	const transactions = useMemo(() => blockResponse?.transactions || [], [blockResponse])
	if (isLoading) {
		return (
			<div className="space-y-6">
				<Skeleton className="h-[600px] w-full" />
			</div>
		)
	}

	if (!transactions) {
		return (
			<div className="flex flex-col items-center justify-center space-y-4 py-12 p-4">
				<h2 className="text-2xl font-bold">Block Not Found</h2>
				<p className="text-muted-foreground">The requested block could not be found.</p>
				<Button asChild>
					<Link to={`/networks/${id}/blocks`}>
						<ArrowLeft className="mr-2 h-4 w-4" />
						Back to Blocks
					</Link>
				</Button>
			</div>
		)
	}

	return (
		<div className="space-y-6 p-4">
			<div className="flex items-center justify-between">
				<div className="space-y-1">
					<h2 className="text-2xl font-bold">Block #{decodedBlock?.number}</h2>
					{/* <p className="text-sm text-muted-foreground">{formatDistanceToNow(new Date(transactions.timestamp || ''), { addSuffix: true })}</p> */}
				</div>
				<Button variant="outline" asChild>
					<Link to={`/networks/${id}/blocks`}>
						<ArrowLeft className="mr-2 h-4 w-4" />
						Back to Blocks
					</Link>
				</Button>
			</div>

			<div className="grid gap-6">
				<Card>
					<CardHeader>
						<CardTitle>Block Information</CardTitle>
					</CardHeader>
					<CardContent>
						<dl className="grid gap-4 sm:grid-cols-2">
							<div>
								<dt className="text-sm font-medium text-muted-foreground">Block Hash</dt>
								<dd className="mt-1 font-mono text-sm break-all">{decodedBlock?.hash}</dd>
							</div>
							<div>
								<dt className="text-sm font-medium text-muted-foreground">Previous Block Hash</dt>
								<dd className="mt-1 font-mono text-sm break-all">{decodedBlock?.previous_hash}</dd>
							</div>
							<div>
								<dt className="text-sm font-medium text-muted-foreground">Timestamp</dt>
								<dd className="mt-1 text-sm">{new Date(decodedBlock?.timestamp || '').toLocaleString()}</dd>
							</div>
							<div>
								<dt className="text-sm font-medium text-muted-foreground">Transactions</dt>
								<dd className="mt-1 text-sm">{transactions.length}</dd>
							</div>
							<div>
								<dt className="text-sm font-medium text-muted-foreground">Data Hash</dt>
								<dd className="mt-1 font-mono text-sm break-all">{decodedBlock?.hash}</dd>
							</div>
						</dl>
					</CardContent>
				</Card>
			</div>
			<Card>
				<CardHeader>
					<CardTitle>Transactions</CardTitle>
				</CardHeader>
				<CardContent>
					{transactions && transactions.length > 0 ? (
						<div className="space-y-4">
							{transactions.map((tx, index: number) => (
								<div key={tx.tx_id || index} className="flex items-center justify-between p-4 border rounded-lg">
									<div className="space-y-1">
										<p className="text-sm font-medium">Transaction ID</p>
										<p className="font-mono text-sm text-muted-foreground break-all">{tx.tx_id}</p>
									</div>
									<DropdownMenu>
										<DropdownMenuTrigger asChild>
											<Button variant="ghost" size="icon">
												<EllipsisVertical className="h-4 w-4" />
											</Button>
										</DropdownMenuTrigger>
										<DropdownMenuContent align="end">
											<DropdownMenuItem>View Details</DropdownMenuItem>
											<DropdownMenuItem>Copy ID</DropdownMenuItem>
										</DropdownMenuContent>
									</DropdownMenu>
								</div>
							))}
						</div>
					) : (
						<div className="text-center py-6">
							<p className="text-sm text-muted-foreground">No transactions found in this block</p>
						</div>
					)}
				</CardContent>
			</Card>
		</div>
	)
}
