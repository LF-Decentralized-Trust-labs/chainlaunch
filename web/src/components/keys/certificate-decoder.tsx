import { X509Certificate } from '@peculiar/x509'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Calendar, Key, LockIcon, Shield, Copy, Download, CheckCircle2, XCircle } from 'lucide-react'
import { format } from 'date-fns'
import { toast } from 'sonner'
import { CertificateThumbprint } from './certificate-thumbprint'

interface CertificateDecoderProps {
	pem: string
}

export function CertificateDecoder({ pem }: CertificateDecoderProps) {
	try {
		const cert = new X509Certificate(pem)
		const now = new Date()
		const isExpired = now > cert.notAfter
		const isNotYetValid = now < cert.notBefore

		let status = 'Valid'
		if (isExpired) status = 'Expired'
		if (isNotYetValid) status = 'Not Yet Valid'

		const getStatusColor = () => {
			if (isExpired) return 'destructive'
			if (isNotYetValid) return 'outline'
			return 'default'
		}

		const getDaysRemaining = () => {
			const days = Math.ceil((cert.notAfter.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
			return days
		}
		console.log('cert', cert)
		const isCA = cert.extensions.some((ext) => ext.type === '2.5.29.19' && (ext as any).ca)

		return (
			<div className="space-y-6">
				<Card className="p-6">
					<div className="grid gap-6">
						<div>
							<div className="flex items-center gap-2 mb-2">
								<Key className="h-4 w-4 text-muted-foreground" />
								<h3 className="text-sm font-medium">Certificate (PEM Format)</h3>
							</div>
							<Card className="p-4 bg-muted/50">
								<div className="flex items-center justify-end mb-2">
									<div className="flex gap-2">
										<Button
											variant="ghost"
											size="sm"
											className="h-8"
											onClick={() => {
												navigator.clipboard.writeText(pem)
												toast.success('Certificate copied to clipboard')
											}}
										>
											<Copy className="h-4 w-4 mr-2" />
											Copy
										</Button>
										<Button
											variant="ghost"
											size="sm"
											className="h-8"
											onClick={() => {
												const blob = new Blob([pem], { type: 'text/plain' })
												const url = window.URL.createObjectURL(blob)
												const a = document.createElement('a')
												a.href = url
												a.download = 'certificate.pem'
												document.body.appendChild(a)
												a.click()
												window.URL.revokeObjectURL(url)
												document.body.removeChild(a)
											}}
										>
											<Download className="h-4 w-4 mr-2" />
											Download
										</Button>
									</div>
								</div>
								<pre className="text-xs font-mono whitespace-pre-wrap overflow-x-auto bg-muted rounded-md p-4">{pem}</pre>
							</Card>
						</div>
						<div>
							<div className="flex items-center justify-between mb-2">
								<div className="flex items-center gap-2">
									<Shield className="h-4 w-4 text-muted-foreground" />
									<h3 className="text-sm font-medium">Certificate Details</h3>
								</div>
								<Badge variant={getStatusColor()}>{status}</Badge>
							</div>
							<Card className="p-4 bg-muted/50">
								<div className="grid gap-4 text-sm">
									<div>
										<span className="text-muted-foreground">Serial Number:</span> <code className="text-xs bg-muted px-2 py-1 rounded font-mono">{cert.serialNumber}</code>
									</div>
									<div>
										<span className="text-muted-foreground">Signature Algorithm:</span> <span className="font-medium">{cert.signatureAlgorithm?.name}</span>
									</div>
								</div>
							</Card>
						</div>

						<div>
							<div className="flex items-center gap-2 mb-2">
								<Key className="h-4 w-4 text-muted-foreground" />
								<h3 className="text-sm font-medium">Public Key</h3>
							</div>
							<Card className="p-4 bg-muted/50">
								<div className="grid gap-4 text-sm">
									<div>
										<span className="text-muted-foreground">Algorithm:</span> <span className="font-medium">{cert.publicKey.algorithm?.name}</span>
									</div>
									{cert.publicKey.algorithm?.name === 'EC' && (
										<div>
											<span className="text-muted-foreground">Curve:</span> <span className="font-medium">{(cert.publicKey as any).namedCurve}</span>
										</div>
									)}
									{cert.publicKey.algorithm?.name === 'RSA' && (
										<div>
											<span className="text-muted-foreground">Key Size:</span> <span className="font-medium">{(cert.publicKey as any).keySize} bits</span>
										</div>
									)}
								</div>
							</Card>
						</div>

						<div>
							<div className="flex items-center gap-2 mb-2">
								<LockIcon className="h-4 w-4 text-muted-foreground" />
								<h3 className="text-sm font-medium">Subject & Issuer</h3>
							</div>
							<div className="grid gap-4">
								<Card className="p-4 bg-muted/50">
									<div className="grid gap-2 text-sm">
										<span className="text-muted-foreground">Subject:</span>
										<pre className="text-xs bg-muted px-2 py-1 rounded block overflow-x-auto">{cert.subject}</pre>
									</div>
								</Card>
								<Card className="p-4 bg-muted/50">
									<div className="grid gap-2 text-sm">
										<span className="text-muted-foreground">Issuer:</span>
										<pre className="text-xs bg-muted px-2 py-1 rounded block overflow-x-auto">{cert.issuer}</pre>
									</div>
								</Card>
							</div>
						</div>

						<div>
							<div className="flex items-center gap-2 mb-2">
								<Calendar className="h-4 w-4 text-muted-foreground" />
								<h3 className="text-sm font-medium">Validity Period</h3>
							</div>
							<Card className="p-4 bg-muted/50">
								<div className="grid gap-4 text-sm">
									<div>
										<span className="text-muted-foreground">Not Before:</span> <span className="font-medium">{format(cert.notBefore, 'PPpp')}</span>
									</div>
									<div>
										<span className="text-muted-foreground">Not After:</span> <span className="font-medium">{format(cert.notAfter, 'PPpp')}</span>
									</div>
								</div>
							</Card>
						</div>

						<div>
							<div className="flex items-center gap-2 mb-2">
								<CheckCircle2 className="h-4 w-4 text-muted-foreground" />
								<h3 className="text-sm font-medium">Certificate Checks</h3>
							</div>
							<Card className="p-4 bg-muted/50">
								<div className="grid gap-4 text-sm">
									<div className="flex items-center justify-between">
										<span className="text-muted-foreground">Valid To</span>
										<div className="flex items-center gap-2">
											{isExpired ? <XCircle className="h-4 w-4 text-destructive" /> : <CheckCircle2 className="h-4 w-4 text-success" />}
											<span>
												{format(cert.notAfter, 'dd MMM yyyy')} ({getDaysRemaining()} days)
											</span>
										</div>
									</div>
									<div className="flex items-center justify-between">
										<span className="text-muted-foreground">Key Size</span>
										<div className="flex items-center gap-2">
											<CheckCircle2 className="h-4 w-4 text-success" />
											<span>
												{cert.publicKey.algorithm?.name} {(cert.publicKey as any).keySize || '256'} bits
											</span>
										</div>
									</div>
									<div className="flex items-center justify-between">
										<span className="text-muted-foreground">Signature Algorithm</span>
										<div className="flex items-center gap-2">
											<CheckCircle2 className="h-4 w-4 text-success" />
											<span>Strong ({cert.signatureAlgorithm?.name})</span>
										</div>
									</div>
								</div>
							</Card>
						</div>

						<div>
							<div className="flex items-center gap-2 mb-2">
								<Shield className="h-4 w-4 text-muted-foreground" />
								<h3 className="text-sm font-medium">Certificate Summary</h3>
							</div>
							<Card className="p-4 bg-muted/50">
								<div className="grid gap-4 text-sm">
									<div>
										<span className="text-muted-foreground">CA Certificate:</span> <span className="font-medium">{isCA ? 'Yes' : 'No'}</span>
									</div>
									<div>
										<span className="text-muted-foreground">Serial Number:</span>{' '}
										<code className="text-xs bg-muted px-2 py-1 rounded font-mono break-all">{cert.serialNumber}</code>
									</div>
									<CertificateThumbprint cert={cert} />
								</div>
							</Card>
						</div>

						{cert.extensions.length > 0 && (
							<div>
								<div className="flex items-center gap-2 mb-2">
									<Shield className="h-4 w-4 text-muted-foreground" />
									<h3 className="text-sm font-medium">Extensions</h3>
								</div>
								<Card className="p-4 bg-muted/50">
									<div className="grid gap-4 text-sm">
										{cert.extensions.map((ext, index) => (
											<div key={index}>
												<div className="font-medium">{ext.type}</div>
												<div className="text-muted-foreground ml-4">Critical: {ext.critical ? 'Yes' : 'No'}</div>
											</div>
										))}
									</div>
								</Card>
							</div>
						)}
					</div>
				</Card>
			</div>
		)
	} catch (error) {
		return (
			<Card className="p-4 bg-destructive/10 text-destructive">
				<p>Failed to decode certificate: {(error as Error).message}</p>
			</Card>
		)
	}
}
