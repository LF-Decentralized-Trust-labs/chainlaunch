import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

export function BackupTargetsSkeleton() {
	return (
		<Card>
			<CardHeader className="flex flex-row items-center justify-between">
				<div>
					<CardTitle>Backup Targets</CardTitle>
					<CardDescription>Configure S3 backup targets for your backups</CardDescription>
				</div>
				<Skeleton className="h-10 w-[120px]" />
			</CardHeader>
			<CardContent>
				<div className="space-y-2">
					<Skeleton className="h-12 w-full" />
					<Skeleton className="h-12 w-full" />
					<Skeleton className="h-12 w-full" />
				</div>
			</CardContent>
		</Card>
	)
} 