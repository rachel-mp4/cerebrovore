<!-- <script lang="ts"> -->
<!--   import { onMount } from "svelte"; -->
<!---->
<!--   interface Props { -->
<!--     onBeforeInput?: (event: InputEvent) => void; -->
<!--     onInputEl?: (el: HTMLTextAreaElement) => void; -->
<!--     imageHandler?: (image: File) => void; -->
<!--     placeholder?: string; -->
<!--     value?: string; -->
<!--     maxlength?: number; -->
<!--     bold?: boolean; -->
<!--     color?: string; -->
<!--     fs?: string; -->
<!--   } -->
<!---->
<!--   let { -->
<!--     onBeforeInput, -->
<!--     placeholder, -->
<!--     value = $bindable(""), -->
<!--     onInputEl, -->
<!--     imageHandler, -->
<!--     maxlength, -->
<!--     bold = false, -->
<!--     color, -->
<!--     fs, -->
<!--   }: Props = $props(); -->
<!---->
<!--   let inputEl: HTMLTextAreaElement; -->
<!--   function adjust(event: Event) { -->
<!--     onInputEl?.(inputEl); -->
<!--   } -->
<!---->
<!--   function bi(event: InputEvent) { -->
<!--     onBeforeInput?.(event); -->
<!--     adjustHeight(); -->
<!--   } -->
<!---->
<!--   function adjustHeight() { -->
<!--     if (inputEl) { -->
<!--       inputEl.style.height = "auto"; -->
<!--       inputEl.style.height = inputEl.scrollHeight + "px"; -->
<!--     } -->
<!--   } -->
<!--   onMount(adjustHeight); -->
<!--   $effect(() => { -->
<!--     value; -->
<!--     adjustHeight(); -->
<!--   }); -->
<!--   const pastifier = (event: ClipboardEvent) => { -->
<!--     const items = event.clipboardData?.items; -->
<!--     if (items === undefined) { -->
<!--       return; -->
<!--     } -->
<!--     for (const item of items) { -->
<!--       if (item.type.startsWith("image/")) { -->
<!--         const blob = item.getAsFile(); -->
<!--         if (blob === null) { -->
<!--           return; -->
<!--         } -->
<!--         event.preventDefault(); -->
<!--         imageHandler?.(blob); -->
<!--       } -->
<!--     } -->
<!--   }; -->
<!-- </script> -->
<!---->
