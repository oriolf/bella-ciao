<script>
  export let type = "info";
  export let content;
</script>

<div class={'alert alert-' + type} role="alert">{content}</div>
