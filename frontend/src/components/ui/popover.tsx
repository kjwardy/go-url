import * as React from 'react';
import { Popover as PopoverPrimitive } from '@base-ui/react/popover';
import { cn } from 'cn';

const Popover = PopoverPrimitive.Root;

function PopoverContent({
  align = 'start',
  side = 'bottom',
  sideOffset = 4,
  anchor,
  className,
  ...props
}: PopoverPrimitive.Popup.Props &
  Pick<
    PopoverPrimitive.Positioner.Props,
    'align' | 'side' | 'sideOffset' | 'anchor'
  >) {
  return (
    <PopoverPrimitive.Portal>
      <PopoverPrimitive.Positioner
        align={align}
        side={side}
        sideOffset={sideOffset}
        anchor={anchor}
        className="z-50 w-(--anchor-width) min-w-64"
      >
        <PopoverPrimitive.Popup
          data-slot="popover-content"
          className={cn(
            'z-50 overflow-hidden rounded-lg border bg-popover text-popover-foreground shadow-md outline-none',
            className,
          )}
          {...props}
        />
      </PopoverPrimitive.Positioner>
    </PopoverPrimitive.Portal>
  );
}

export { Popover, PopoverContent };
