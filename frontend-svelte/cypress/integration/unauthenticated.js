describe('Non authenticated user', () => {
    beforeEach(() => {
        cy.visit('/', { onBeforeLoad: () => { localStorage.setItem('bella-ciao.lang', 'en'); } });
    });

    it('shows the log in card', () => {
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
        cy.contains('h1', 'questions');
    });
});