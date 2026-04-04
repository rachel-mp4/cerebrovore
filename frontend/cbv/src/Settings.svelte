<script lang="ts">
  import VolumeSettings from "./lib/components/VolumeSettings.svelte";
  const ntv = localStorage.getItem("new-threads");
  let wantsNewThreads = $state(ntv !== "");
  $effect(() => {
    localStorage.setItem("new-threads", wantsNewThreads ? "yes" : "");
  });
</script>

<input type="checkbox" bind:checked={wantsNewThreads} id="wants-new-threads" />
<label for="wants-new-threads"
  >notify locally whenever new threads are created? (requires refresh to go into
  effect)</label
>
<VolumeSettings />
