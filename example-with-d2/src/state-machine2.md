## STATE MACHINE

```d2
over: Game Over {shape: circle}
running: Game Running {shape: circle}
pause: Game Paused {shape: circle}

over -> running : Click Start
pause -> running : Click Resume
running -> pause : Click Pause
running -> over : Die
```

Next two examples kudos to [https://venilnoronha.io/a-simple-state-machine-framework-in-go](https://venilnoronha.io/a-simple-state-machine-framework-in-go)

Like a light switch

```d2
light.on -> light.off : Switch off
light.off -> light.on : Switch on
```

Order processing example

```d2
Order Creating -> Order Failed : Fail
Order Failed -> Order Create : Create
Order Creating -> Order Placed : Place
Order Placed -> Charging Card : Charge
Charging Card -> Transaction Failed : Fail
Transaction Failed -> Charging Card : Charge
Charging Card -> Order Shipped : Ship
```
