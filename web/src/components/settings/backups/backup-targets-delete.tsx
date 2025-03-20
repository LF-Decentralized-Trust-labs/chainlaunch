import { type HttpBackupTargetResponse } from '@/api/client'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from '@/components/ui/alert-dialog'
import { DropdownMenuItem } from '@/components/ui/dropdown-menu'
import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import { deleteBackupsTargetsById } from '@/api/client'

interface BackupTargetsDeleteProps {
	target: HttpBackupTargetResponse
	onSuccess: () => void
}

export function BackupTargetsDelete({ target, onSuccess }: BackupTargetsDeleteProps) {
	const deleteMutation = useMutation({
		mutationFn: async () => {
			try {
				await deleteBackupsTargetsById({ path: { id: target.id! } })
			} catch (error: any) {
				if (error.status === 500) {
					throw new Error('Internal server error. Please try again later.')
				}
				throw error
			}
		},
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
					<AlertDialogDescription>
						This will permanently delete the backup target. This action cannot be undone.
					</AlertDialogDescription>
				</AlertDialogHeader>
				<AlertDialogFooter>
					<AlertDialogCancel>Cancel</AlertDialogCancel>
					<AlertDialogAction variant="destructive" onClick={() => deleteMutation.mutate()}>
						Delete
					</AlertDialogAction>
				</AlertDialogFooter>
			</AlertDialogContent>
		</AlertDialog>
	)
} 