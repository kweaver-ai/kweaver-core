import React from 'react';
import { shallow } from 'enzyme';

import Container from '../Container';

const init = (props: any) => shallow(<Container {...props} />);

describe('Container', () => {
  it('class name containerRoot is exists', () => {
    const wrapper = init({});

    expect(wrapper.find('.containerRoot').exists()).toBe(true);
  });
});
