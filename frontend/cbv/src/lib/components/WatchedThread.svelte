<script lang="ts">
  import { b36encodenumber } from "../utils";
  import type { WatchThread } from "../types";
  interface Props {
    item: WatchThread;
  }
  let { item }: Props = $props();
</script>

<div class="watched-thread">
  <div class="bump-count{item.bumpLimit ? ' bump-limit' : ''}">
    {item.bumpLimit ? "-" : b36encodenumber(item.bumps)}
  </div>
  <div class="thread-link">
    <a href="/t/{b36encodenumber(item.id)}">
      {item.topic ?? `#${b36encodenumber(item.id)}`}
    </a>
  </div>
</div>
