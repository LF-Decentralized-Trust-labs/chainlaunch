import { Card } from '../ui/card'
import { Badge } from '../ui/badge'
import { Shield } from 'lucide-react'
import { ValidatorItem } from './validator-item'
import { Skeleton } from '../ui/skeleton'

interface ValidatorListProps {
  validatorIds: number[]
}

export function ValidatorList({ validatorIds }: ValidatorListProps) {


  return (
    <div className="space-y-2">
      {validatorIds.map((keyId, index) => (
        <ValidatorItem key={keyId} keyId={keyId} index={index} />
      ))}
    </div>
  )
} 