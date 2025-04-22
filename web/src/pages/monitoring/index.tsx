import { NotificationProvidersList } from '@/components/settings/notifications/notification-providers-list'

export default function MonitoringPage() {
	return (
		<div className="container space-y-6 p-4">
			<div>
				<h1 className="text-2xl font-semibold tracking-tight">Monitoring</h1>
				<p className="text-sm text-muted-foreground">Configure monitoring and notification settings</p>
			</div>

			<NotificationProvidersList />
		</div>
	)
}
