<script lang="ts">
  import { b36encodenumber } from "../utils";
  import type { WatchThread } from "../types";
  interface Props {
    item: WatchThread;
    idx: number;
    watchIdx: (idx: number) => void;
    rmIdx: (idx: number) => void;
  }
  let { item, idx, watchIdx, rmIdx }: Props = $props();
</script>

<div class="new-thread">
  <button
    class="ignore-thread"
    onclick={() => {
      rmIdx(idx);
    }}
  >
    [x]
  </button>
  <button
    class="watch-thread"
    onclick={() => {
      watchIdx(idx);
    }}
  >
    {item.topic ?? `#${b36encodenumber(item.id)}`}
  </button>
</div>
