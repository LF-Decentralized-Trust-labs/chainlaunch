import { type HttpBackupTargetResponse } from '@/api/client'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { EllipsisVertical } from 'lucide-react'
import { BackupTargetsUpdate } from './backup-targets-update'
import { BackupTargetsDelete } from './backup-targets-delete'

interface BackupTargetsTableProps {
	targets: HttpBackupTargetResponse[]
	onSuccess: () => void
}

export function BackupTargetsTable({ targets, onSuccess }: BackupTargetsTableProps) {
	return (
		<div className="max-h-[calc(100vh-20rem)] overflow-y-auto">
			<Table>
				<TableHeader>
					<TableRow>
						<TableHead>Name</TableHead>
						<TableHead>Endpoint</TableHead>
						<TableHead>Bucket</TableHead>
						<TableHead>Region</TableHead>
						<TableHead className="w-[50px]"></TableHead>
					</TableRow>
				</TableHeader>
				<TableBody>
					{targets.map((target) => (
						<TableRow key={target.id}>
							<TableCell>{target.name}</TableCell>
							<TableCell>{target.endpoint}</TableCell>
							<TableCell>{target.bucketName}</TableCell>
							<TableCell>{target.region}</TableCell>
							<TableCell>
								<DropdownMenu>
									<DropdownMenuTrigger asChild>
										<Button variant="ghost" size="icon">
											<EllipsisVertical className="h-4 w-4" />
										</Button>
									</DropdownMenuTrigger>
									<DropdownMenuContent align="end">
										<BackupTargetsUpdate target={target} onSuccess={onSuccess} />
										<BackupTargetsDelete target={target} onSuccess={onSuccess} />
									</DropdownMenuContent>
								</DropdownMenu>
							</TableCell>
						</TableRow>
					))}
				</TableBody>
			</Table>
		</div>
	)
} 