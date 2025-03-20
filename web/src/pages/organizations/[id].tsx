import { getOrganizationsByIdOptions } from '@/api/client/@tanstack/react-query.gen'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { CertificateViewer } from '@/components/ui/certificate-viewer'
import { Skeleton } from '@/components/ui/skeleton'
import { TimeAgo } from '@/components/ui/time-ago'
import { useQuery } from '@tanstack/react-query'
import { ArrowLeft, Building2, Key as KeyIcon } from 'lucide-react'
import { Link, useParams } from 'react-router-dom'

export default function OrganizationDetailPage() {
	const { id } = useParams()
	const { data: org, isLoading } = useQuery({
		...getOrganizationsByIdOptions({
			path: { id: Number(id) },
		}),
	})

	if (isLoading) {
		return (
			<div className="flex-1 p-8">
				<div className="max-w-4xl mx-auto">
					<div className="mb-8">
						<Skeleton className="h-8 w-32 mb-2" />
						<Skeleton className="h-5 w-64" />
					</div>
					<div className="space-y-8">
						<Card className="p-6">
							<div className="space-y-4">
								<div className="flex items-center gap-4">
									<Skeleton className="h-12 w-12 rounded-lg" />
									<div>
										<Skeleton className="h-6 w-48 mb-2" />
										<Skeleton className="h-4 w-32" />
									</div>
								</div>
								<Skeleton className="h-24 w-full" />
							</div>
						</Card>
					</div>
				</div>
			</div>
		)
	}

	if (!org) {
		return (
			<div className="flex-1 p-8">
				<div className="max-w-4xl mx-auto text-center">
					<Building2 className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
					<h1 className="text-2xl font-semibold mb-2">Organization not found</h1>
					<p className="text-muted-foreground mb-8">The organization you're looking for doesn't exist or you don't have access to it.</p>
					<Button asChild>
						<Link to="/fabric/organizations">
							<ArrowLeft className="mr-2 h-4 w-4" />
							Back to Organizations
						</Link>
					</Button>
				</div>
			</div>
		)
	}

	return (
		<div className="flex-1 p-8">
			<div className="max-w-4xl mx-auto">
				<div className="flex items-center gap-2 text-muted-foreground mb-8">
					<Button variant="ghost" size="sm" asChild>
						<Link to="/fabric/organizations">
							<ArrowLeft className="mr-2 h-4 w-4" />
							Organizations
						</Link>
					</Button>
				</div>

				<div className="flex items-center justify-between mb-8">
					<div>
						<h1 className="text-2xl font-semibold mb-1">{org.mspId}</h1>
						<p className="text-muted-foreground">
							Created <TimeAgo date={org.createdAt!} />
						</p>
					</div>
				</div>

				<div className="space-y-8">
					{/* Organization Info Card */}
					<Card className="p-6">
						<div className="flex items-center gap-4 mb-6">
							<div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center">
								<Building2 className="h-6 w-6 text-primary" />
							</div>
							<div>
								<h2 className="text-lg font-semibold">Organization Information</h2>
								<p className="text-sm text-muted-foreground">Details about your organization</p>
							</div>
						</div>

						<div className="space-y-6">
							<div>
								<h3 className="text-sm font-medium mb-2">MSP ID</h3>
								<p className="text-sm text-muted-foreground">{org.mspId}</p>
							</div>

							{org.description && (
								<div>
									<h3 className="text-sm font-medium mb-2">Description</h3>
									<p className="text-sm text-muted-foreground">{org.description}</p>
								</div>
							)}
						</div>
					</Card>

					<Card className="p-4">
						<div className="flex items-center justify-between">
							<div>
								<h3 className="font-medium mb-1">Sign Certificate</h3>
								<p className="text-sm text-muted-foreground">Organization signing certificate</p>
							</div>
							<Badge variant="outline">Active</Badge>
						</div>
						<div className="mt-4">
							<p className="text-xs text-muted-foreground mb-1">Certificate</p>
							<CertificateViewer certificate={org.signCertificate!} label="Sign Certificate" className="w-full" />
						</div>
						<div className="mt-4">
							<p className="text-xs text-muted-foreground mb-1">Public Key</p>
							<pre className="text-sm font-mono bg-muted p-4 rounded-lg overflow-x-auto whitespace-pre-wrap break-all">{org.signPublicKey}</pre>
						</div>
					</Card>

					{/* TLS Certificate */}
					<Card className="p-4">
						<div className="flex items-center justify-between">
							<div>
								<h3 className="font-medium mb-1">TLS Certificate</h3>
								<p className="text-sm text-muted-foreground">Organization TLS certificate</p>
							</div>
							<Badge variant="outline">Active</Badge>
						</div>
						<div className="mt-4">
							<p className="text-xs text-muted-foreground mb-1">Certificate</p>
							<CertificateViewer certificate={org.tlsCertificate!} label="TLS Certificate" className="w-full" />
						</div>
						<div className="mt-4">
							<p className="text-xs text-muted-foreground mb-1">Public Key</p>
							<pre className="text-sm font-mono bg-muted p-4 rounded-lg overflow-x-auto whitespace-pre-wrap break-all">{org.tlsPublicKey}</pre>
						</div>
					</Card>
				</div>
			</div>
		</div>
	)
}
