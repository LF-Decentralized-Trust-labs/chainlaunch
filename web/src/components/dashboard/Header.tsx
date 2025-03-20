import { ThemeToggle } from '@/components/theme/ThemeToggle'
import { useBreadcrumbs } from '@/contexts/BreadcrumbContext'
import { Link } from 'react-router-dom'
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbPage, BreadcrumbSeparator } from '../ui/breadcrumb'
import { Separator } from '../ui/separator'
import { SidebarTrigger } from '../ui/sidebar'

export function Header() {
	const { breadcrumbs } = useBreadcrumbs()

	return (
		<header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
			<div className="flex justify-between w-full">
				<div className="flex items-center">
					<SidebarTrigger className="-ml-1" />
					<Separator orientation="vertical" className="mr-2 h-4" />
					<Breadcrumb>
						<BreadcrumbList>
							{breadcrumbs.map((item, index) => (
								<BreadcrumbItem key={index} className={index === 0 ? '' : ''}>
									{index < breadcrumbs.length - 1 ? (
										<>
											<BreadcrumbLink asChild href={item.href ?? '#'}>
												<Link to={item.href ?? '#'}>{item.label}</Link>
											</BreadcrumbLink>
											<BreadcrumbSeparator className="" />
										</>
									) : (
										<BreadcrumbPage>{item.label}</BreadcrumbPage>
									)}
								</BreadcrumbItem>
							))}
						</BreadcrumbList>
					</Breadcrumb>
				</div>
				<div className="ml-auto flex items-center space-x-4">
					<ThemeToggle />
				</div>
			</div>
		</header>
	)
}
