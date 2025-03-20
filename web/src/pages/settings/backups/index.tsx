import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { BackupTargets } from '@/components/settings/backups/backup-targets'
import { BackupSchedules } from '@/components/settings/backups/backup-schedules'
import { BackupsList } from '@/components/settings/backups/backups-list'
import { useSearchParams, useNavigate } from 'react-router-dom'

export default function BackupsPage() {
	const navigate = useNavigate()
	const [searchParams] = useSearchParams()
	const tab = searchParams.get('tab') || 'targets'

	const handleTabChange = (value: string) => {
		navigate(`/settings/backups?tab=${value}`, {
			replace: true
		})
	}

	return (
		<div className="container space-y-6">
			<div>
				<h1 className="text-2xl font-semibold tracking-tight">Backups</h1>
				<p className="text-sm text-muted-foreground">Manage your backup targets, schedules and create backups</p>
			</div>

			<Tabs value={tab} onValueChange={handleTabChange} className="space-y-4">
				<TabsList>
					<TabsTrigger value="targets">Backup Targets</TabsTrigger>
					<TabsTrigger value="schedules">Backup Schedules</TabsTrigger>
					<TabsTrigger value="backups">Backups</TabsTrigger>
				</TabsList>

				<TabsContent value="targets" className="space-y-4">
					<BackupTargets />
				</TabsContent>

				<TabsContent value="schedules" className="space-y-4">
					<BackupSchedules />
				</TabsContent>

				<TabsContent value="backups" className="space-y-4">
					<BackupsList />
				</TabsContent>
			</Tabs>
		</div>
	)
}
