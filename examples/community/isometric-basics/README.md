# Isometric view basics

Created by [Sergio Vera](https://github.com/svera).

Isometric view is a display method used to create an illusion of 3D for an otherwise 2D game - sometimes referred to as pseudo 3D or 2.5D.

Implementing an isometric view can be done in many ways, but for the sake of simplicity we'll implement a tile-based approach, which is the most efficient and widely used method.

In the tile-based approach, each visual element is broken down into smaller pieces, called tiles, of a standard size. These tiles will be arranged to form the game world according to pre-determined level data - usually a 2D array.

For a detailed explanation about the maths behind this, read [http://clintbellanger.net/articles/isometric_math/](http://clintbellanger.net/articles/isometric_math/).

![Result](result.png)
