declare module '*.gif' {
  const src: string;
  export default src;
}

declare module '*.jpg' {
  const src: string;
  export default src;
}

declare module '*.jpeg' {
  const src: string;
  export default src;
}

declare module '*.png' {
  const src: string;
  export default src;
}

declare module '*.svg' {
  import React from 'react';

  const ReactComponent: React.FunctionComponent<
    React.SVGProps<SVGSVGElement> & { className?: string; style?: React.CSSProperties }
  >;

  export default ReactComponent;
}

declare module '*.css' {
  const classes: { readonly [key: string]: string };
  export default classes;
  export = classes;
}

declare module '*.less' {
  const classes: { [key: string]: string };
  export default classes;
  export = classes;
}

declare module 'swagger-client/*';
declare module 'remarkable*';
declare module 'randombytes';
declare module 'css.escape';
declare module 'sha.js';
declare module 'react-syntax-highlighter/*';
declare module 'xml-but-prettier';
declare module '@aishu-tech/components/dist/dip-components.min';
