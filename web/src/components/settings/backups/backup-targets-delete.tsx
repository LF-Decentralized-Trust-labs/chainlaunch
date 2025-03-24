import { type HttpBackupTargetResponse } from '@/api/client'
import { deleteBackupsTargetsByIdMutation } from '@/api/client/@tanstack/react-query.gen'
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
	AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { DropdownMenuItem } from '@/components/ui/dropdown-menu'
import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'

interface BackupTargetsDeleteProps {
	target: HttpBackupTargetResponse
	onSuccess: () => void
}

export function BackupTargetsDelete({ target, onSuccess }: BackupTargetsDeleteProps) {
	const deleteMutation = useMutation({
		...deleteBackupsTargetsByIdMutation(),

		onSuccess: () => {
			toast.success('Backup target deleted successfully')
			onSuccess()
		},
		onError: (error) => {
			toast.error('Failed to delete backup target', {
				description: error.message,
			})
		},
	})

	return (
		<AlertDialog>
			<AlertDialogTrigger asChild>
				<DropdownMenuItem className="text-destructive" onSelect={(e) => e.preventDefault()}>
					Delete
				</DropdownMenuItem>
			</AlertDialogTrigger>
			<AlertDialogContent>
				<AlertDialogHeader>
					<AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
					<AlertDialogDescription>This will permanently delete the backup target. This action cannot be undone.</AlertDialogDescription>
				</AlertDialogHeader>
				<AlertDialogFooter>
					<AlertDialogCancel>Cancel</AlertDialogCancel>
					<AlertDialogAction onClick={() => deleteMutation.mutate({ path: { id: target.id! } })}>Delete</AlertDialogAction>
				</AlertDialogFooter>
			</AlertDialogContent>
		</AlertDialog>
	)
}
