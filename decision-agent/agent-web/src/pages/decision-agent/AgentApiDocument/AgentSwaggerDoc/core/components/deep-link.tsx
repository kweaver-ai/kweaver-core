import PropTypes from 'prop-types';

export const DeepLink = ({ enabled, path, text }) => {
  const { origin } = window.location;
  return (
    <a className="nostyle" onClick={enabled ? e => e.preventDefault() : null} href={enabled ? `#/${path}` : null}>
      <span>
        {origin}
        {text}
      </span>
    </a>
  );
};
DeepLink.propTypes = {
  enabled: PropTypes.bool,
  isShown: PropTypes.bool,
  path: PropTypes.string,
  text: PropTypes.node,
};

export default DeepLink;
