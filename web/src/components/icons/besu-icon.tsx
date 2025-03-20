import besuLogo from '../../../public/blockchains/besu_favicon.svg'

interface BesuIconProps {
  className?: string
}

export function BesuIcon({ className }: BesuIconProps) {
  return <img src={besuLogo} alt="Besu" className={className} />
} 