describe('Non authenticated user', () => {
	function setEnglish() {
		// TODO get this to work properly, aparently it only works at the start of the test, and later changes again
		Object.defineProperty(window.navigator, 'language', { value: 'en' });
	}

	beforeEach(() => {
		cy.visit('/', {
			onBeforeLoad: setEnglish(),
			onLoad: setEnglish()
		})
	});

	it('shows the log in card', () => {
		console.log(window.navigator.language);
		cy.contains('.card-header', 'Log in')
	});

	it('navigates to /candidates', () => {
		cy.get('nav a').contains('Candidates').click();
		cy.url().should('include', '/candidates');
		cy.contains('.card-header', 'candidate 1');
	});

	it('navigates to /faq', () => {
		cy.get('nav a').contains('FAQ').click();
		cy.url().should('include', '/faq');
		cy.contains('h1', 's');
	});
});