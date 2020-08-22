<script>
  export let segment;
</script>

<style>
nav.navbar {
	margin-bottom: 10px;
}
</style>

<nav class="navbar navbar-expand-lg sticky-top navbar-dark bg-primary">
  <a class="navbar-brand" href=".">Home</a>
  <button class="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbar" aria-controls="navbar" aria-expanded="false" aria-label="Toggle navigation">
    <span class="navbar-toggler-icon"></span>
  </button>
  <div class="collapse navbar-collapse" id="navbar">
    <ul class="navbar-nav mr-auto">
      <li class="{segment === 'faq' ? 'nav-item active' : 'nav-item'}">
        <a class="nav-link" aria-current={segment === 'faq' ? 'page' : undefined} href="faq">
		  FAQ
		  {#if segment === 'faq'}
		  <span class="sr-only">(current)</span>
		  {/if}
        </a>
      </li>
    </ul>
  </div>
</nav>
