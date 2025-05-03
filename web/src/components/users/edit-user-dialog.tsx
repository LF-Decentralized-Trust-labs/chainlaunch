import { AuthUpdateUserRequest, AuthUserResponse } from '@/api/client/types.gen'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

const editUserFormSchema = z.object({
	role: z.enum(['admin', 'manager', 'viewer'], {
		required_error: 'Please select a role',
	}),
})

type EditUserFormValues = z.infer<typeof editUserFormSchema>

interface EditUserDialogProps {
	user: AuthUserResponse
	onSubmit: (data: AuthUpdateUserRequest) => Promise<void>
	isLoading?: boolean
	trigger?: React.ReactNode
}

export function EditUserDialog({ user, onSubmit, isLoading, trigger }: EditUserDialogProps) {
	const [open, setOpen] = useState(false)

	const form = useForm<EditUserFormValues>({
		resolver: zodResolver(editUserFormSchema),
		defaultValues: {
			role: user.role || 'viewer',
		},
	})

	const handleSubmit = async (data: EditUserFormValues) => {
		await onSubmit({ role: data.role })
		setOpen(false)
	}

	return (
		<Dialog open={open} onOpenChange={setOpen}>
			<DialogTrigger asChild>{trigger || <Button variant="ghost">Edit User</Button>}</DialogTrigger>
			<DialogContent className="sm:max-w-[425px]">
				<DialogHeader>
					<DialogTitle>Edit User</DialogTitle>
					<DialogDescription>Update role for user "{user.username}".</DialogDescription>
				</DialogHeader>
				<Form {...form}>
					<form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
						<FormField
							control={form.control}
							name="role"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Role</FormLabel>
									<Select onValueChange={field.onChange} defaultValue={field.value}>
										<FormControl>
											<SelectTrigger>
												<SelectValue placeholder="Select a role" />
											</SelectTrigger>
										</FormControl>
										<SelectContent>
											<SelectItem value="admin">Admin</SelectItem>
											<SelectItem value="manager">Manager</SelectItem>
											<SelectItem value="viewer">Viewer</SelectItem>
										</SelectContent>
									</Select>
									<FormMessage />
								</FormItem>
							)}
						/>

						<Button type="submit" className="w-full" disabled={isLoading}>
							{isLoading ? 'Updating...' : 'Update User'}
						</Button>
					</form>
				</Form>
			</DialogContent>
		</Dialog>
	)
}
