'use strict';

export default (d3, force, tick) => {
  function dragstart(d, i) {
    force.stop(); // stops the force auto positioning before you start dragging
  };

  function dragmove(d, i) {
    d.px += d3.event.dx;
    d.py += d3.event.dy;
    d.x += d3.event.dx;
    d.y += d3.event.dy;
    tick(); // this is the key to make it work together with updating both px,py,x,y on d !
  };

  function dragend(d, i) {
      d.fixed = true; // of course set the node to fixed so the force doesn't include the node in its auto positioning stuff
      tick();
      force.resume();
  };

  return d3.behavior.drag()
      .on('dragstart', dragstart)
      .on('drag', dragmove)
      .on('dragend', dragend);
};
