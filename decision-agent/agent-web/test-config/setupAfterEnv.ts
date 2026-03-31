import { cleanup } from '@testing-library/react';
import '@testing-library/jest-dom/jest-globals';
import '@testing-library/jest-dom';

afterEach(() => {
  cleanup();
  document.body.innerHTML = '';
});

jest.mock('query-string', () => ({
  stringify: jest.fn,
}));
