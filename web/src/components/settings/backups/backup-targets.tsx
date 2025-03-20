import { useState } from 'react'
import { useMutation, useQueryClient, useQuery } from '@tanstack/react-query'
import { deleteBackupsTargetsById, postBackupsTargets, type HttpBackupTargetResponse, getBackupsTargets } from '@/api/client'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage, FormDescription } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { zodResolver } from '@hookform/resolvers/zod'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { EllipsisVertical, Plus } from 'lucide-react'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Skeleton } from '@/components/ui/skeleton'
import { toast } from 'sonner'
import { getBackupsTargetsOptions } from '@/api/client/@tanstack/react-query.gen'
import { Checkbox } from '@/components/ui/checkbox'
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
import { BackupTargetsCreate } from './backup-targets-create'
import { BackupTargetsUpdate } from './backup-targets-update'
import { BackupTargetsDelete } from './backup-targets-delete'
import { BackupTargetsSkeleton } from './backup-targets-skeleton'
import { BackupTargetsEmpty } from './backup-targets-empty'
import { BackupTargetsTable } from './backup-targets-table'

const targetFormSchema = z.object({
	name: z.string().min(1),
	endpoint: z.string().min(1),
	type: z.literal('S3'),
	accessKeyId: z.string().min(1),
	secretKey: z.string().min(1),
	bucketName: z.string().min(1),
	bucketPath: z.string().min(1),
	region: z.string().min(1),
	forcePathStyle: z.boolean().optional(),
})

type TargetFormValues = z.infer<typeof targetFormSchema>

export function BackupTargets() {
	const {
		data: targets = [],
		isLoading,
		refetch,
	} = useQuery({
		...getBackupsTargetsOptions(),
	})
	console.log(targets, typeof targets)
	const [open, setOpen] = useState(false)

	const form = useForm<TargetFormValues>({
		resolver: zodResolver(targetFormSchema),
		defaultValues: {
			type: 'S3',
			forcePathStyle: false,
		},
	})

	const createMutation = useMutation({
		mutationFn: async (values: TargetFormValues) => {
			try {
				await postBackupsTargets({ body: values })
			} catch (error: any) {
				if (error.status === 500) {
					throw new Error('Internal server error. Please try again later.')
				}
				throw error
			}
		},
		onSuccess: () => {
			toast.success('Backup target created successfully')
			setOpen(false)
			form.reset()
			refetch()
		},
		onError: (error) => {
			toast.error('Failed to create backup target', {
				description: error.message,
			})
		},
	})

	const deleteMutation = useMutation({
		mutationFn: async (id: number) => {
			try {
				await deleteBackupsTargetsById({ path: { id } })
			} catch (error: any) {
				if (error.status === 500) {
					throw new Error('Internal server error. Please try again later.')
				}
				throw error
			}
		},
		onSuccess: () => {
			toast.success('Backup target deleted successfully')
			refetch()
		},
		onError: (error) => {
			toast.error('Failed to delete backup target', {
				description: error.message,
			})
		},
	})

	if (isLoading) {
		return <BackupTargetsSkeleton />
	}

	return (
		<Card>
			<CardHeader className="flex flex-row items-center justify-between">
				<div>
					<CardTitle>Backup Targets</CardTitle>
					<CardDescription>Configure S3 backup targets for your backups</CardDescription>
				</div>
				<BackupTargetsCreate onSuccess={refetch} />
			</CardHeader>
			<CardContent>{targets.length === 0 ? <BackupTargetsEmpty /> : <BackupTargetsTable targets={targets} onSuccess={refetch} />}</CardContent>
		</Card>
	)
}
