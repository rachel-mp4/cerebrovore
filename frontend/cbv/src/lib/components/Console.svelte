<script lang="ts">
  import type { LogItem } from "../types";
  import { numToHex } from "../colors";
  interface Props {
    log: Array<LogItem>;
  }
  let { log }: Props = $props();

  const randPosition = (l: LogItem): string => {
    const top =
      Math.abs((999.999 * Math.sin(l.id * l.id * 11.11)) % 1) * 95 +
      4 * (Math.sin(l.time * 7.7) % 1);
    const left =
      Math.abs((999.999 * Math.sin(l.id * l.id * 22.22)) % 1) * 90 +
      5 * (Math.sin(l.time * 14.14) % 1);
    return `top: ${top}%; left: ${left}%`;
  };
</script>

{#each log as logitem (logitem.key)}
  <span
    style={randPosition(logitem)}
    style:--accent={numToHex(logitem.color ?? 0)}
    class="logitem {logitem.type}"
  >
    0x{logitem.binary}
  </span>
{/each}
