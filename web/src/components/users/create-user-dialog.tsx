import { AuthCreateUserRequest } from '@/api/client/types.gen'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { UserPlus } from 'lucide-react'
import { useState } from 'react'
import { UserForm } from './user-form'

interface CreateUserDialogProps {
  onSubmit: (data: AuthCreateUserRequest) => void
  isLoading?: boolean
}

export function CreateUserDialog({ onSubmit, isLoading }: CreateUserDialogProps) {
  const [open, setOpen] = useState(false)

  const handleSubmit = async (data: AuthCreateUserRequest) => {
    await onSubmit(data)
    setOpen(false)
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <UserPlus className="mr-2 h-4 w-4" />
          Add User
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Create New User</DialogTitle>
          <DialogDescription>
            Add a new user to the system. They will be able to log in with these credentials.
          </DialogDescription>
        </DialogHeader>
        <UserForm onSubmit={handleSubmit} isLoading={isLoading} />
      </DialogContent>
    </Dialog>
  )
} 