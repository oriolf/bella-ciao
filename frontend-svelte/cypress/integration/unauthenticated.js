describe('Non authenticated user', () => {
    function setEnglish() {
        localStorage.setItem('bella-ciao.lang', 'en');
    }

    beforeEach(() => {
        cy.visit('/', { onBeforeLoad: setEnglish() });
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
        cy.contains('h1', 's');
    });
});