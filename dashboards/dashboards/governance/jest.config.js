export default {
    testEnvironment: 'jsdom',
    transform: {
        '^.+\\.js$': 'babel-jest',
    },
    moduleNameMapper: {
        '\\.(css|less|scss|sass)$': '<rootDir>/tests/__mocks__/styleMock.js',
    },
    collectCoverageFrom: [
        '**/*.js',
        '!**/node_modules/**',
        '!**/tests/**',
        '!**/coverage/**',
    ],
    coverageThreshold: {
        global: {
            branches: 75,
            functions: 75,
            lines: 75,
            statements: 75,
        },
    },
};
