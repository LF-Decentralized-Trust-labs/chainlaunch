import { useQuery } from '@tanstack/react-query'
import { Card } from '../ui/card'
import { Badge } from '../ui/badge'
import { Shield, AlertCircle } from 'lucide-react'
import { Skeleton } from '../ui/skeleton'
import { Alert, AlertDescription } from '../ui/alert'
import { getKeysByIdOptions } from '@/api/client/@tanstack/react-query.gen'

interface ValidatorItemProps {
	keyId: number
	index: number
}

export function ValidatorItem({ keyId, index }: ValidatorItemProps) {
	const {
		data: validatorKey,
		isLoading,
		error,
	} = useQuery({
		...getKeysByIdOptions({
			path: { id: keyId },
		}),
	})

	if (isLoading) {
		return (
			<Card className="p-3">
				<div className="flex items-center gap-2">
					<Skeleton className="h-6 w-6 rounded-full" />
					<div className="flex flex-col gap-1">
						<Skeleton className="h-4 w-24" />
						<Skeleton className="h-4 w-48" />
					</div>
				</div>
			</Card>
		)
	}

	if (error) {
		return (
			<Alert variant="destructive">
				<AlertCircle className="h-4 w-4" />
				<AlertDescription>Failed to load validator {index + 1} data</AlertDescription>
			</Alert>
		)
	}

	return (
		<Card className="p-3">
			<div className="flex items-center gap-2">
				<Badge variant="secondary" className="h-6 w-6 rounded-full p-1">
					<Shield className="h-4 w-4" />
				</Badge>
				<div className="flex flex-col gap-1">
					<div className="text-xs text-muted-foreground">Validator {index + 1}</div>
					<code className="text-xs">{validatorKey?.ethereumAddress}</code>
					{validatorKey && <code className="text-xs text-muted-foreground">Key: {validatorKey.publicKey?.slice(0, 20)}...</code>}
				</div>
			</div>
		</Card>
	)
}
