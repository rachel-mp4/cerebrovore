<script lang="ts">
  interface Props {
    src: string;
    alt: string | undefined;
    gifoverride?: boolean;
  }
  let { src, alt, gifoverride }: Props = $props();
  let isgif = $derived(src.endsWith(".gif") || gifoverride);
</script>

{#if isgif}
  <div class="image-wrapper">
    <img class="bg-img" {src} {alt} /><img class="fg-img" {src} {alt} />
  </div>
{:else}
  <div class="image-wrapper thumb" data-thumb="{src}&thumb=yes" data-full={src}>
    <img class="bg-img" src="{src}&thumb=yes" {alt} /><img
      class="fg-img"
      src="{src}&thumb=yes"
      {alt}
    />
  </div>
{/if}
