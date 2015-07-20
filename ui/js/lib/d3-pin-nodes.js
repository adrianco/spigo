'use strict';

export default (d3, force, tick) => {
  const dragstart = (d, i) => {
    force.stop();
  };

  const dragmove = (d, i) => {
    d.px += d3.event.dx;
    d.py += d3.event.dy;
    d.x += d3.event.dx;
    d.y += d3.event.dy;
    tick();
  };

  const dragend = (d, i) => {
      d.fixed = true;
      tick();
      force.resume();
  };

  return d3.behavior.drag()
      .on('dragstart', dragstart)
      .on('drag', dragmove)
      .on('dragend', dragend);
};
