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
  <div class="image-wrapper thumb">
    <img class="bg-img" {src} {alt} title={alt} /><img
      class="fg-img"
      {src}
      {alt}
      title={alt}
      onload={() => {
        document.dispatchEvent(new CustomEvent("lrc:scrollIfAttached"));
      }}
    />
  </div>
{:else}
  <div class="image-wrapper thumb" data-thumb="{src}&thumb=yes" data-full={src}>
    <img class="bg-img" src="{src}&thumb=yes" {alt} title={alt} /><img
      class="fg-img"
      src="{src}&thumb=yes"
      {alt}
      title={alt}
      onload={() => {
        document.dispatchEvent(new CustomEvent("lrc:scrollIfAttached"));
      }}
    />
  </div>
{/if}
