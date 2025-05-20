/**
 * Silly blinky
 */

module Blinky (
    input clk,
    input rst,
    output reg blink
);

  reg [31:0] MEM[0:7];
  reg [2:0] counter = 0;

  initial begin
    blink = 0;
    MEM[0] = 32'hde;
    MEM[1] = 32'had;
    MEM[2] = 32'hbe;
    MEM[3] = 32'hef;
    MEM[4] = 32'h55;
    MEM[5] = 32'haa;
    MEM[6] = 32'h55;
    MEM[7] = 32'haa;
  end

  always @(posedge clk) begin
    counter <= counter + 1;
    MEM[counter] <= {{29{1'b0}}, counter};
    if (counter == 0) begin
      blink = ~blink;
    end
  end

endmodule
